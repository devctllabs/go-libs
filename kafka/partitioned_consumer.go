package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// PartitionedConsumerConfig configures partition-isolated processing.
type PartitionedConsumerConfig struct {
	Consumer ConsumerConfig
	// MaxConcurrentPartitions limits active partition handlers. Zero is
	// unlimited.
	MaxConcurrentPartitions int
}

func (config PartitionedConsumerConfig) validate() error {
	if err := config.Consumer.validate(); err != nil {
		return err
	}
	if config.MaxConcurrentPartitions < 0 {
		return errors.New("kafka: maximum concurrent partitions must not be negative")
	}
	return nil
}

// PartitionedConsumer processes different partitions concurrently while
// preserving sequential handling within each partition.
type PartitionedConsumer[T any] struct {
	config  PartitionedConsumerConfig
	client  consumerClient
	decoder Decoder[T]
	factory PartitionHandlerFactory[T]

	mu      sync.Mutex
	started bool
}

// NewPartitionedConsumer constructs a single-use partition-isolated consumer.
func NewPartitionedConsumer[T any](
	config PartitionedConsumerConfig,
	decoder Decoder[T],
	factory PartitionHandlerFactory[T],
) (*PartitionedConsumer[T], error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	if decoder == nil {
		return nil, errors.New("kafka: decoder must not be nil")
	}
	if factory == nil {
		return nil, errors.New("kafka: partition handler factory must not be nil")
	}
	consumerConfig := config.Consumer
	options := []kgo.Opt{
		kgo.SeedBrokers(consumerConfig.Brokers...),
		kgo.ConsumerGroup(consumerConfig.Group),
		kgo.ConsumeTopics(consumerConfig.Topics...),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
		kgo.RebalanceTimeout(consumerConfig.RebalanceTimeout),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	}
	if hooks := observerHooks(consumerConfig.Observer, consumerConfig.Group); len(hooks) > 0 {
		options = append(options, kgo.WithHooks(hooks...))
	}
	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("kafka: create partitioned consumer client: %w", err)
	}
	return newPartitionedConsumerWithClient(config, client, decoder, factory), nil
}

// Run polls and processes records until cancellation or a terminal partition
// failure. It owns the client and all factory-created handlers.
func (consumer *PartitionedConsumer[T]) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("kafka: context must not be nil")
	}
	consumer.mu.Lock()
	if consumer.started {
		consumer.mu.Unlock()
		return ErrAlreadyRun
	}
	consumer.started = true
	consumer.mu.Unlock()
	defer consumer.client.Close()

	handlers := make(map[TopicPartition]PartitionHandler[T])
	defer func() { consumer.closeHandlers(ctx, handlers) }()
	serializedClient := &serializedCommitClient{consumerClient: consumer.client}
	pending := make(map[TopicPartition][]*kgo.Record)
	flushAt := make(map[TopicPartition]time.Time)
	for ctx.Err() == nil {
		records, err := consumer.pollRecords(ctx, flushAt)
		if err != nil {
			return err
		}
		for partition, records := range groupByPartition(records) {
			if len(pending[partition]) == 0 {
				flushAt[partition] = time.Now().Add(consumer.config.Consumer.Batch.FlushInterval)
			}
			pending[partition] = append(pending[partition], records...)
		}
		ready := readyPartitions(pending, flushAt, consumer.config.Consumer.Batch.MaxSize, time.Now())
		if len(ready) == 0 {
			continue
		}
		if err := consumer.processPartitions(ctx, ready, handlers, serializedClient); err != nil {
			return err
		}
		for partition := range ready {
			delete(pending, partition)
			delete(flushAt, partition)
		}
		if len(pending) == 0 {
			consumer.client.AllowRebalance()
		}
	}
	if len(pending) > 0 {
		drainCtx, cancelDrain := context.WithTimeout(
			context.WithoutCancel(ctx),
			consumer.config.Consumer.ShutdownTimeout,
		)
		defer cancelDrain()
		if err := consumer.processPartitions(drainCtx, pending, handlers, serializedClient); err != nil {
			return fmt.Errorf("kafka: drain partition batches: %w", err)
		}
		consumer.client.AllowRebalance()
	}
	return cleanCancellation(ctx)
}

func (consumer *PartitionedConsumer[T]) pollRecords(
	ctx context.Context,
	flushAt map[TopicPartition]time.Time,
) ([]*kgo.Record, error) {
	pollCtx := ctx
	cancelPoll := func() {}
	if deadline, ok := earliestDeadline(flushAt); ok {
		pollCtx, cancelPoll = context.WithDeadline(ctx, deadline)
	}
	fetches := consumer.client.PollRecords(pollCtx, consumer.config.Consumer.Batch.MaxSize)
	flushExpired := errors.Is(pollCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil
	cancelPoll()
	if err := fetches.Err(); err != nil {
		if flushExpired {
			return nil, nil
		}
		if ctx.Err() != nil || errors.Is(err, context.Canceled) || fetches.IsClientClosed() {
			return nil, nil
		}
		return nil, fmt.Errorf("kafka: poll partition records: %w", err)
	}
	return fetches.Records(), nil
}

func earliestDeadline(deadlines map[TopicPartition]time.Time) (time.Time, bool) {
	var earliest time.Time
	for _, deadline := range deadlines {
		if earliest.IsZero() || deadline.Before(earliest) {
			earliest = deadline
		}
	}
	return earliest, !earliest.IsZero()
}

func readyPartitions(
	pending map[TopicPartition][]*kgo.Record,
	deadlines map[TopicPartition]time.Time,
	maxSize int,
	now time.Time,
) map[TopicPartition][]*kgo.Record {
	ready := make(map[TopicPartition][]*kgo.Record)
	for partition, records := range pending {
		if len(records) >= maxSize || !deadlines[partition].After(now) {
			ready[partition] = records
		}
	}
	return ready
}

func (consumer *PartitionedConsumer[T]) processPartitions(
	ctx context.Context,
	grouped map[TopicPartition][]*kgo.Record,
	handlers map[TopicPartition]PartitionHandler[T],
	client consumerClient,
) error {
	limit := consumer.config.MaxConcurrentPartitions
	if limit == 0 || limit > len(grouped) {
		limit = len(grouped)
	}
	semaphore := make(chan struct{}, limit)
	errorsChannel := make(chan error, len(grouped))
	var group sync.WaitGroup
	for partition, records := range grouped {
		handler, ok := handlers[partition]
		if !ok {
			var err error
			handler, err = consumer.factory.NewHandler(ctx, partition)
			if err != nil {
				return fmt.Errorf("kafka: create handler for %s/%d: %w", partition.Topic, partition.Partition, err)
			}
			if handler == nil {
				return fmt.Errorf("kafka: factory returned nil handler for %s/%d", partition.Topic, partition.Partition)
			}
			handlers[partition] = handler
		}
		group.Add(1)
		go func(records []*kgo.Record, handler PartitionHandler[T]) {
			defer group.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			processor := &Consumer[T]{
				config: consumer.config.Consumer, client: client,
				decoder: consumer.decoder, handler: handler,
			}
			maxSize := consumer.config.Consumer.Batch.MaxSize
			for start := 0; start < len(records); start += maxSize {
				end := min(start+maxSize, len(records))
				if err := processor.processBatch(ctx, records[start:end]); err != nil {
					errorsChannel <- err
					return
				}
			}
		}(records, handler)
	}
	group.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		return fmt.Errorf("kafka: process partition: %w", err)
	}
	return nil
}

func (consumer *PartitionedConsumer[T]) closeHandlers(
	ctx context.Context,
	handlers map[TopicPartition]PartitionHandler[T],
) {
	closeCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		consumer.config.Consumer.ShutdownTimeout,
	)
	defer cancel()
	for _, handler := range handlers {
		_ = handler.Close(closeCtx)
	}
}

func groupByPartition(records []*kgo.Record) map[TopicPartition][]*kgo.Record {
	grouped := make(map[TopicPartition][]*kgo.Record)
	for _, record := range records {
		partition := TopicPartition{Topic: record.Topic, Partition: record.Partition}
		grouped[partition] = append(grouped[partition], record)
	}
	return grouped
}

type serializedCommitClient struct {
	consumerClient
	mu sync.Mutex
}

func (client *serializedCommitClient) CommitRecords(ctx context.Context, records ...*kgo.Record) error {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.consumerClient.CommitRecords(ctx, records...)
}

func newPartitionedConsumerWithClient[T any](
	config PartitionedConsumerConfig,
	client consumerClient,
	decoder Decoder[T],
	factory PartitionHandlerFactory[T],
) *PartitionedConsumer[T] {
	return &PartitionedConsumer[T]{config: config, client: client, decoder: decoder, factory: factory}
}
