package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	libretry "github.com/devctllabs/go-libs/retry"
	"github.com/twmb/franz-go/pkg/kgo"
)

// ErrAlreadyRun reports a second Run call on a single-use consumer.
var ErrAlreadyRun = errors.New("kafka: consumer has already run")

// Consumer processes mixed-partition batches with one handler instance.
type Consumer[T any] struct {
	config  ConsumerConfig
	client  consumerClient
	decoder Decoder[T]
	handler BatchHandler[T]

	mu      sync.Mutex
	started bool
}

// NewConsumer constructs a mixed-partition consumer. The returned consumer is
// single-use; Run owns and closes its franz-go client.
func NewConsumer[T any](
	config ConsumerConfig,
	decoder Decoder[T],
	handler BatchHandler[T],
) (*Consumer[T], error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	if decoder == nil {
		return nil, errors.New("kafka: decoder must not be nil")
	}
	if handler == nil {
		return nil, errors.New("kafka: batch handler must not be nil")
	}
	options := []kgo.Opt{
		kgo.SeedBrokers(config.Brokers...),
		kgo.ConsumerGroup(config.Group),
		kgo.ConsumeTopics(config.Topics...),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
		kgo.RebalanceTimeout(config.RebalanceTimeout),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	}
	if hooks := observerHooks(config.Observer, config.Group); len(hooks) > 0 {
		options = append(options, kgo.WithHooks(hooks...))
	}
	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("kafka: create consumer client: %w", err)
	}
	return newConsumerWithClient(config, client, decoder, handler), nil
}

// Run polls, processes, and commits batches until ctx is canceled or a
// terminal failure occurs. A plain cancellation is a clean stop.
func (consumer *Consumer[T]) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("kafka: context must not be nil")
	}
	if err := consumer.beginRun(); err != nil {
		return err
	}
	defer consumer.client.Close()

	raw := make([]*kgo.Record, 0, consumer.config.Batch.MaxSize)
	var flushAt time.Time
	for ctx.Err() == nil {
		pollCtx := ctx
		cancelPoll := func() {}
		if !flushAt.IsZero() {
			pollCtx, cancelPoll = context.WithDeadline(ctx, flushAt)
		}
		fetches := consumer.client.PollRecords(pollCtx, consumer.config.Batch.MaxSize-len(raw))
		flushExpired := errors.Is(pollCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil
		cancelPoll()
		if err := fetches.Err(); err != nil {
			if ctx.Err() != nil || errors.Is(err, context.Canceled) || fetches.IsClientClosed() {
				return cleanCancellation(ctx)
			}
			return fmt.Errorf("kafka: poll records: %w", err)
		}
		wasEmpty := len(raw) == 0
		raw = append(raw, fetches.Records()...)
		if wasEmpty && len(raw) > 0 && consumer.config.Batch.FlushInterval > 0 {
			flushAt = time.Now().Add(consumer.config.Batch.FlushInterval)
		}
		if len(raw) < consumer.config.Batch.MaxSize && !flushExpired {
			continue
		}
		if len(raw) == 0 {
			flushAt = time.Time{}
			continue
		}
		if err := consumer.processBatch(ctx, raw); err != nil {
			return err
		}
		raw = raw[:0]
		flushAt = time.Time{}
		consumer.client.AllowRebalance()
	}
	if len(raw) > 0 {
		drainCtx, cancelDrain := context.WithTimeout(context.WithoutCancel(ctx), consumer.config.ShutdownTimeout)
		defer cancelDrain()
		if err := consumer.processBatch(drainCtx, raw); err != nil {
			return fmt.Errorf("kafka: drain active batch: %w", err)
		}
		consumer.client.AllowRebalance()
	}
	return cleanCancellation(ctx)
}

func (consumer *Consumer[T]) beginRun() error {
	consumer.mu.Lock()
	defer consumer.mu.Unlock()
	if consumer.started {
		return ErrAlreadyRun
	}
	consumer.started = true
	return nil
}

func (consumer *Consumer[T]) processBatch(ctx context.Context, records []*kgo.Record) error {
	decoded, err := consumer.decodeRecords(ctx, records)
	if err != nil {
		return err
	}
	if len(decoded.messages) > 0 {
		batch := newBatch(decoded.messages)
		if err := consumer.handleBatch(ctx, batch, decoded.records); err != nil {
			return fmt.Errorf("kafka: handle batch: %w", err)
		}
		handlerRejections, err := consumer.applyRejectPolicy(batch, decoded.records)
		if err != nil {
			return err
		}
		decoded.dispositions = append(decoded.dispositions, handlerRejections...)
		if consumer.config.OnReject == RejectDLQ {
			decoded.rejections = append(decoded.rejections, handlerRejections...)
		}
		countDecisions(&decoded.result, batch.decisionsBuffer)
	}
	dlqDelivered, err := consumer.deliverDLQ(ctx, decoded.rejections)
	if err != nil {
		return fmt.Errorf("kafka: deliver rejected records: %w", err)
	}
	observer := effectiveObserver(consumer.config.Observer)
	if err := doWithRetry(ctx, effectiveRetry(consumer.config.Retry, consumer.config.CommitRetry), observer, AttemptCommit, records, func(commitCtx context.Context) error {
		return consumer.client.CommitRecords(commitCtx, records...)
	}); err != nil {
		return fmt.Errorf("kafka: commit batch: %w", err)
	}
	consumer.observeCompletedBatch(ctx, observer, decoded, dlqDelivered)
	return nil
}

type decodedBatch[T any] struct {
	messages     []Message[T]
	records      []*kgo.Record
	rejections   []recordRejection
	dispositions []recordRejection
	result       BatchResult
}

func (consumer *Consumer[T]) decodeRecords(
	ctx context.Context,
	records []*kgo.Record,
) (*decodedBatch[T], error) {
	decoded := &decodedBatch[T]{
		messages: make([]Message[T], 0, len(records)),
		records:  make([]*kgo.Record, 0, len(records)),
		result:   BatchResult{Size: len(records)},
	}
	for index, record := range records {
		value, err := consumer.decoder.Decode(ctx, record.Value)
		if err != nil {
			decoded.result.Rejected++
			rejection := recordRejection{
				record: record,
				kind:   "decode",
				cause:  &DecodeError{Index: index, Err: err},
			}
			switch consumer.config.OnReject {
			case RejectStop:
				return nil, rejectedMessageError(record, rejection.cause)
			case RejectDLQ:
				decoded.rejections = append(decoded.rejections, rejection)
			}
			decoded.dispositions = append(decoded.dispositions, rejection)
			continue
		}
		decoded.messages = append(decoded.messages, decodedMessage(record, value))
		decoded.records = append(decoded.records, record)
	}
	return decoded, nil
}

func countDecisions(result *BatchResult, decisions []messageDecision) {
	for _, decision := range decisions {
		switch decision.outcome {
		case messageProcessed:
			result.Processed++
		case messageSkipped:
			result.Skipped++
		case messageRejected:
			result.Rejected++
		}
	}
}

func (consumer *Consumer[T]) observeCompletedBatch(
	ctx context.Context,
	observer Observer,
	decoded *decodedBatch[T],
	dlqDelivered bool,
) {
	if dlqDelivered {
		decoded.result.DLQ = len(decoded.rejections)
	}
	decoded.result.Dropped = decoded.result.Rejected - decoded.result.DLQ
	for _, rejection := range decoded.dispositions {
		kind := DispositionDropped
		if consumer.config.OnReject == RejectDLQ && dlqDelivered {
			kind = DispositionDLQ
		}
		observer.RecordDisposition(ctx, RecordDisposition{
			Record: recordMetadata(rejection.record),
			Kind:   kind,
			Cause:  rejection.cause,
		})
	}
	observer.BatchCompleted(ctx, decoded.result)
}

func effectiveRetry(baseline RetryConfig, override *RetryConfig) RetryConfig {
	if override != nil {
		return *override
	}
	return baseline
}

func (consumer *Consumer[T]) applyRejectPolicy(
	batch *Batch[T],
	records []*kgo.Record,
) ([]recordRejection, error) {
	rejections := make([]recordRejection, 0)
	for index, decision := range batch.decisionsBuffer {
		if decision.outcome != messageRejected {
			continue
		}
		if consumer.config.OnReject == RejectStop {
			return nil, rejectedMessageError(records[index], decision.err)
		}
		rejections = append(rejections, recordRejection{
			record: records[index],
			kind:   "handler",
			cause:  decision.err,
		})
	}
	return rejections, nil
}

func rejectedMessageError(record *kgo.Record, cause error) *RejectedMessageError {
	return &RejectedMessageError{
		Topic:     record.Topic,
		Partition: record.Partition,
		Offset:    record.Offset,
		Cause:     cause,
	}
}

func (consumer *Consumer[T]) handleBatch(
	ctx context.Context,
	batch *Batch[T],
	records []*kgo.Record,
) error {
	return doWithRetry(ctx, consumer.config.Retry, effectiveObserver(consumer.config.Observer), AttemptHandler, records, func(attemptCtx context.Context) error {
		batch.reset()
		return consumer.handler.Handle(attemptCtx, batch)
	})
}

func doWithRetry(
	ctx context.Context,
	config RetryConfig,
	observer Observer,
	phase AttemptPhase,
	records []*kgo.Record,
	operation libretry.Operation,
) error {
	attempt := uint(0)
	observedOperation := func(attemptCtx context.Context) error {
		attempt++
		operationCtx, done := observer.StartAttempt(attemptCtx, Attempt{
			Phase: phase, Attempt: attempt, Records: metadata(records),
		})
		if done == nil {
			done = func(error) {}
		}
		err := operation(operationCtx)
		done(err)
		return err
	}
	if config.MaxAttempts == 1 {
		return observedOperation(ctx)
	}
	options := make([]libretry.Option, 0, 3)
	if config.MaxAttempts > 0 {
		options = append(options, libretry.WithMaxAttempts(config.MaxAttempts))
	}
	if config.MaxElapsedTime > 0 {
		options = append(options, libretry.WithMaxElapsedTime(config.MaxElapsedTime))
	}
	options = append(options, libretry.WithNotify(func(attempt uint, err error, nextDelay time.Duration) {
		observer.Retry(ctx, RetryEvent{Phase: phase, Attempt: attempt, NextDelay: nextDelay, Err: err})
	}))
	return libretry.Do(ctx, config.Policy, observedOperation, options...)
}

func decodedMessage[T any](record *kgo.Record, value T) Message[T] {
	var headers []Header
	if record.Headers != nil {
		headers = make([]Header, len(record.Headers))
		for index, header := range record.Headers {
			headers[index] = Header{Key: header.Key, Value: header.Value}
		}
	}
	return Message[T]{
		Topic:     record.Topic,
		Partition: record.Partition,
		Offset:    record.Offset,
		Timestamp: record.Timestamp,
		Key:       record.Key,
		Headers:   headers,
		Value:     value,
	}
}

func cleanCancellation(ctx context.Context) error {
	cause := context.Cause(ctx)
	if cause == nil || cause == context.Canceled {
		return nil
	}
	return cause
}

func newConsumerWithClient[T any](
	config ConsumerConfig,
	client consumerClient,
	decoder Decoder[T],
	handler BatchHandler[T],
) *Consumer[T] {
	return &Consumer[T]{config: config, client: client, decoder: decoder, handler: handler}
}
