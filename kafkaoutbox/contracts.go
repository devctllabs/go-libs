package kafkaoutbox

import (
	"context"

	"github.com/devctllabs/go-libs/kafka"
)

//go:generate go tool mockgen -destination=contracts.gen_test.go -package=kafkaoutbox -typed . Appender,BatchPublisher,Observer
//go:generate go tool mockgen -destination=encoder.gen_test.go -package=kafkaoutbox -typed github.com/devctllabs/go-libs/kafka Encoder

// Event describes one domain event to persist in an outbox.
type Event[T any] struct {
	Topic         string
	AggregateType string
	AggregateID   string
	Type          string
	Value         T
}

// Appender persists encoded events in the caller's active business transaction.
type Appender interface {
	// Append persists event atomically with the transaction carried by ctx.
	Append(ctx context.Context, event Event[[]byte]) error
}

// BatchPublisher delivers encoded outbox messages to Kafka.
type BatchPublisher interface {
	// SendBatch sends all messages and waits for their broker delivery results.
	SendBatch(ctx context.Context, messages []kafka.OutgoingMessage[[]byte]) error
}
