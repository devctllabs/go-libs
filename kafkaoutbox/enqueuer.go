package kafkaoutbox

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devctllabs/go-libs/kafka"
)

// EnqueuerConfig configures enqueue observation.
type EnqueuerConfig struct {
	Observer Observer
}

// Enqueuer serializes typed events before appending them to an outbox.
type Enqueuer[T any] struct {
	appender Appender
	encoder  kafka.Encoder[T]
	observer Observer
}

// NewEnqueuer constructs a typed transactional outbox enqueuer.
func NewEnqueuer[T any](
	config EnqueuerConfig,
	appender Appender,
	encoder kafka.Encoder[T],
) (*Enqueuer[T], error) {
	if appender == nil {
		return nil, errors.New("kafkaoutbox: appender must not be nil")
	}
	if encoder == nil {
		return nil, errors.New("kafkaoutbox: encoder must not be nil")
	}
	return &Enqueuer[T]{
		appender: appender,
		encoder:  encoder,
		observer: effectiveObserver(config.Observer),
	}, nil
}

// Enqueue encodes event and appends it using the transaction carried by ctx.
func (enqueuer *Enqueuer[T]) Enqueue(ctx context.Context, event Event[T]) (resultErr error) {
	if err := validateEvent(ctx, event); err != nil {
		return err
	}
	ctx, done := enqueuer.observer.StartOperation(ctx, Operation{Phase: OperationEnqueue})
	if done == nil {
		done = func(error) {}
	}
	defer func() { done(resultErr) }()
	payload, err := enqueuer.encoder.Encode(ctx, event.Value)
	if err != nil {
		return fmt.Errorf("kafkaoutbox: encoder.Encode: %w", err)
	}
	if payload == nil {
		payload = []byte{}
	}
	if err := enqueuer.appender.Append(ctx, Event[[]byte]{
		Topic:         event.Topic,
		AggregateType: event.AggregateType,
		AggregateID:   event.AggregateID,
		Type:          event.Type,
		Value:         payload,
	}); err != nil {
		return fmt.Errorf("kafkaoutbox: appender.Append: %w", err)
	}
	enqueuer.observer.Enqueued(ctx)
	return nil
}

func validateEvent[T any](ctx context.Context, event Event[T]) error {
	if ctx == nil {
		return errors.New("kafkaoutbox: context must not be nil")
	}
	fields := []struct {
		name  string
		value string
	}{
		{name: "topic", value: event.Topic},
		{name: "aggregate type", value: event.AggregateType},
		{name: "aggregate id", value: event.AggregateID},
		{name: "event type", value: event.Type},
	}
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("kafkaoutbox: %s must not be blank", field.name)
		}
	}
	if strings.Contains(event.AggregateType, "::") || strings.Contains(event.AggregateID, "::") {
		return errors.New("kafkaoutbox: aggregate type and id must not contain the reserved delimiter")
	}
	return nil
}
