package kafkaoutbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/devctllabs/go-libs/postgresdb"
	"github.com/google/uuid"
	"github.com/zeebo/xxh3"
)

// ErrTransactionRequired reports an Append call without an active business transaction.
var ErrTransactionRequired = errors.New("kafkaoutbox: active transaction required")

// PollingStore appends events and supplies polling worker storage operations.
type PollingStore struct {
	endpoint *postgresdb.Endpoint
}

// NewPollingStore constructs a transaction-aware polling outbox store.
func NewPollingStore(endpoint *postgresdb.Endpoint) (*PollingStore, error) {
	if endpoint == nil {
		return nil, errors.New("kafkaoutbox: PostgreSQL endpoint must not be nil")
	}
	return &PollingStore{endpoint: endpoint}, nil
}

// Append persists event in the active business transaction carried by ctx.
func (store *PollingStore) Append(ctx context.Context, event Event[[]byte]) error {
	if !store.endpoint.InTransaction(ctx) {
		return ErrTransactionRequired
	}
	if err := validateEvent(ctx, event); err != nil {
		return err
	}
	id, aggregateKey, payload, err := prepareStoredEvent(event)
	if err != nil {
		return err
	}
	traceContext := traceContextFrom(ctx)
	_, err = store.endpoint.Exec(ctx, `
		INSERT INTO outbox_events (
			id, topic, aggregatetype, aggregateid, aggregatekey, type, payload, routing_hash,
			traceparent, tracestate
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, event.Topic, event.AggregateType, event.AggregateID, aggregateKey, event.Type, payload,
		int64(xxh3.HashString(aggregateKey)), traceContext.traceparent, traceContext.tracestate)
	if err != nil {
		return fmt.Errorf("kafkaoutbox: endpoint.Exec: %w", err)
	}
	return nil
}

func prepareStoredEvent(event Event[[]byte]) (uuid.UUID, string, []byte, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, "", nil, fmt.Errorf("kafkaoutbox: uuid.NewV7: %w", err)
	}
	payload := event.Value
	if payload == nil {
		payload = []byte{}
	}
	return id, event.AggregateType + "::" + event.AggregateID, payload, nil
}

var _ Appender = (*PollingStore)(nil)
