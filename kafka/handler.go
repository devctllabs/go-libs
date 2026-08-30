package kafka

import "context"

// BatchHandler processes one immutable batch. A non-nil error retries the
// whole batch and discards per-message decisions from that attempt.
type BatchHandler[T any] interface {
	// Handle processes batch synchronously and must stop work when ctx is
	// canceled. Implementations must not retain batch data after returning.
	Handle(ctx context.Context, batch *Batch[T]) error
}

// BatchHandlerFunc adapts a function to BatchHandler.
type BatchHandlerFunc[T any] func(ctx context.Context, batch *Batch[T]) error

// Handle calls handler(ctx, batch).
func (handler BatchHandlerFunc[T]) Handle(ctx context.Context, batch *Batch[T]) error {
	return handler(ctx, batch)
}

// TopicPartition uniquely identifies one Kafka partition.
type TopicPartition struct {
	Topic     string
	Partition int32
}

// PartitionHandler is owned by one assigned partition and is never called
// concurrently by PartitionedConsumer.
type PartitionHandler[T any] interface {
	BatchHandler[T]
	// Close releases partition-scoped resources after processing stops.
	Close(ctx context.Context) error
}

// PartitionHandlerFactory creates one handler for each observed partition.
// Implementations must be concurrency-safe.
type PartitionHandlerFactory[T any] interface {
	// NewHandler constructs the handler owned by partition.
	NewHandler(ctx context.Context, partition TopicPartition) (PartitionHandler[T], error)
}

// PartitionHandlerFactoryFunc adapts a function to PartitionHandlerFactory.
type PartitionHandlerFactoryFunc[T any] func(
	ctx context.Context,
	partition TopicPartition,
) (PartitionHandler[T], error)

// NewHandler calls factory(ctx, partition).
func (factory PartitionHandlerFactoryFunc[T]) NewHandler(
	ctx context.Context,
	partition TopicPartition,
) (PartitionHandler[T], error) {
	return factory(ctx, partition)
}
