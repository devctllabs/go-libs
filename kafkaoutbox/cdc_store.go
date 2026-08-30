package kafkaoutbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/devctllabs/go-libs/postgresdb"
)

// CDCStore appends events to an append-only Debezium outbox table.
type CDCStore struct {
	endpoint *postgresdb.Endpoint
}

// NewCDCStore constructs a transaction-aware Debezium outbox store.
func NewCDCStore(endpoint *postgresdb.Endpoint) (*CDCStore, error) {
	if endpoint == nil {
		return nil, errors.New("kafkaoutbox: PostgreSQL endpoint must not be nil")
	}
	return &CDCStore{endpoint: endpoint}, nil
}

// Append persists event in the active business transaction carried by ctx.
func (store *CDCStore) Append(ctx context.Context, event Event[[]byte]) error {
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
			id, topic, aggregatetype, aggregateid, aggregatekey, type, payload, tracingspancontext
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, event.Topic, event.AggregateType, event.AggregateID, aggregateKey, event.Type, payload,
		traceContext.debeziumProperties())
	if err != nil {
		return fmt.Errorf("kafkaoutbox: endpoint.Exec: %w", err)
	}
	return nil
}

var _ Appender = (*CDCStore)(nil)
