package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

// ErrClosed reports an operation attempted after a runtime began closing.
var ErrClosed = errors.New("kafka: runtime is closed")

// ProducerConfig contains the required producer connection values.
type ProducerConfig struct {
	Brokers []string
	// Observer may provide standard franz-go hooks. Producer operations do not
	// emit consumer lifecycle callbacks.
	Observer Observer
}

// OutgoingMessage is one typed record to produce.
type OutgoingMessage[T any] struct {
	Topic   string
	Key     []byte
	Headers []Header
	Value   T
}

// EncodeError reports a local encoding failure before any batch records were
// sent. Index identifies the input message.
type EncodeError struct {
	Index int
	Err   error
}

// DeliveryFailure describes one failed record from SendBatch. Index refers to
// the caller's input slice; payload data is never retained.
type DeliveryFailure struct {
	Index     int
	Topic     string
	Partition int32
	Err       error
}

// BatchDeliveryError reports partial or complete broker delivery failure.
type BatchDeliveryError struct {
	failures  []DeliveryFailure
	succeeded int
	cause     error
}

// Error implements error.
func (err *BatchDeliveryError) Error() string {
	return fmt.Sprintf(
		"kafka: deliver batch: %d succeeded, %d failed",
		err.succeeded,
		len(err.failures),
	)
}

// Unwrap returns all retained delivery failures as one error chain.
func (err *BatchDeliveryError) Unwrap() error {
	return err.cause
}

// Failures returns a copy of the per-record failure metadata.
func (err *BatchDeliveryError) Failures() []DeliveryFailure {
	return append([]DeliveryFailure(nil), err.failures...)
}

// Succeeded returns the number of records acknowledged without error.
func (err *BatchDeliveryError) Succeeded() int {
	return err.succeeded
}

// Error implements error.
func (err *EncodeError) Error() string {
	return fmt.Sprintf("kafka: encode message %d: %v", err.Index, err.Err)
}

// Unwrap returns the encoder failure.
func (err *EncodeError) Unwrap() error {
	return err.Err
}

// Producer synchronously publishes typed Kafka messages.
type Producer[T any] struct {
	client  recordProducer
	encoder Encoder[T]

	mu     sync.Mutex
	closed bool
	active sync.WaitGroup
	close  sync.Once
}

// NewProducer constructs an instance-owned producer without opening a network
// connection. The first send establishes broker connections lazily.
func NewProducer[T any](config ProducerConfig, encoder Encoder[T]) (*Producer[T], error) {
	if err := validateStrings("broker", config.Brokers); err != nil {
		return nil, err
	}
	if encoder == nil {
		return nil, errors.New("kafka: encoder must not be nil")
	}
	options := []kgo.Opt{
		kgo.SeedBrokers(config.Brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	}
	if hooks := observerHooks(config.Observer, ""); len(hooks) > 0 {
		options = append(options, kgo.WithHooks(hooks...))
	}
	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("kafka: create producer client: %w", err)
	}
	return newProducerWithClient(client, encoder), nil
}

// Send encodes message and waits for its broker delivery result.
func (producer *Producer[T]) Send(ctx context.Context, message OutgoingMessage[T]) error {
	return producer.SendBatch(ctx, []OutgoingMessage[T]{message})
}

// SendBatch encodes all messages before sending any and waits for every broker
// delivery result.
func (producer *Producer[T]) SendBatch(ctx context.Context, messages []OutgoingMessage[T]) error {
	if ctx == nil {
		return errors.New("kafka: context must not be nil")
	}
	if err := producer.beginSend(); err != nil {
		return err
	}
	defer producer.active.Done()

	records, err := producer.encode(ctx, messages)
	if err != nil || len(records) == 0 {
		return err
	}
	return deliveryError(producer.client.ProduceSync(ctx, records...))
}

// Close prevents new sends, waits for active sends, and closes the owned
// franz-go client. It is safe to call concurrently and repeatedly.
func (producer *Producer[T]) Close(ctx context.Context) error {
	if ctx == nil {
		return errors.New("kafka: context must not be nil")
	}
	producer.mu.Lock()
	producer.closed = true
	producer.mu.Unlock()

	done := make(chan struct{})
	go func() {
		producer.active.Wait()
		close(done)
	}()
	select {
	case <-ctx.Done():
		producer.close.Do(producer.client.Close)
		return ctx.Err()
	case <-done:
		producer.close.Do(producer.client.Close)
		return nil
	}
}

func (producer *Producer[T]) beginSend() error {
	producer.mu.Lock()
	defer producer.mu.Unlock()
	if producer.closed {
		return ErrClosed
	}
	producer.active.Add(1)
	return nil
}

func (producer *Producer[T]) encode(ctx context.Context, messages []OutgoingMessage[T]) ([]*kgo.Record, error) {
	records := make([]*kgo.Record, len(messages))
	for index, message := range messages {
		if strings.TrimSpace(message.Topic) == "" {
			return nil, fmt.Errorf("kafka: message %d topic must not be blank", index)
		}
		value, err := producer.encoder.Encode(ctx, message.Value)
		if err != nil {
			return nil, &EncodeError{Index: index, Err: err}
		}
		records[index] = &kgo.Record{
			Topic:   message.Topic,
			Key:     message.Key,
			Value:   value,
			Headers: recordHeaders(message.Headers),
		}
	}
	return records, nil
}

func recordHeaders(headers []Header) []kgo.RecordHeader {
	converted := make([]kgo.RecordHeader, len(headers))
	for index, header := range headers {
		converted[index] = kgo.RecordHeader{Key: header.Key, Value: header.Value}
	}
	return converted
}

func deliveryError(results kgo.ProduceResults) error {
	failures := make([]DeliveryFailure, 0)
	causes := make([]error, 0)
	for index, result := range results {
		if result.Err == nil {
			continue
		}
		failures = append(failures, DeliveryFailure{
			Index:     index,
			Topic:     result.Record.Topic,
			Partition: result.Record.Partition,
			Err:       result.Err,
		})
		causes = append(causes, result.Err)
	}
	if len(failures) == 0 {
		return nil
	}
	return &BatchDeliveryError{
		failures:  failures,
		succeeded: len(results) - len(failures),
		cause:     errors.Join(causes...),
	}
}

func newProducerWithClient[T any](client recordProducer, encoder Encoder[T]) *Producer[T] {
	return &Producer[T]{client: client, encoder: encoder}
}
