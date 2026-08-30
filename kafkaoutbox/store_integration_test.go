//go:build integration

package kafkaoutbox_test

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"sort"
	"testing"

	"github.com/devctllabs/go-libs/kafkaoutbox"
	"github.com/devctllabs/go-libs/postgresdb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/zeebo/xxh3"
	"go.opentelemetry.io/otel/trace"
)

func TestPollingStoreAppendsOnlyInsideBusinessTransaction(t *testing.T) {
	ctx := context.Background()
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)
	event := kafkaoutbox.Event[[]byte]{
		Topic:         "orders",
		AggregateType: "Order",
		AggregateID:   "42",
		Type:          "OrderPaid",
		Value:         []byte("wire"),
	}

	err = store.Append(ctx, event)
	require.ErrorIs(t, err, kafkaoutbox.ErrTransactionRequired)
	require.Equal(t, 0, eventCount(t, ctx, db.Writer()))

	err = db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		return store.Append(txCtx, event)
	})
	require.NoError(t, err)
	require.Equal(t, 1, eventCount(t, ctx, db.Writer()))
	assertStoredEvent(t, ctx, db.Writer())

	rollbackErr := errors.New("rollback")
	err = db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, store.Append(txCtx, event))
		return rollbackErr
	})
	require.ErrorIs(t, err, rollbackErr)
	require.Equal(t, 1, eventCount(t, ctx, db.Writer()))
}

func TestPollingStorePersistsW3CTraceContext(t *testing.T) {
	ctx := context.Background()
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)
	traceState, err := trace.ParseTraceState("vendor=value")
	require.NoError(t, err)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SpanID:     trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
		TraceFlags: trace.FlagsSampled,
		TraceState: traceState,
	})
	tracedCtx := trace.ContextWithSpanContext(ctx, spanContext)

	require.NoError(t, db.Writer().WithinTx(tracedCtx, func(txCtx context.Context) error {
		return store.Append(txCtx, kafkaoutbox.Event[[]byte]{
			Topic: "orders", AggregateType: "Order", AggregateID: "42", Type: "OrderPaid",
		})
	}))

	var traceparent, tracestate string
	require.NoError(t, db.Writer().QueryRow(ctx, `
		SELECT traceparent, tracestate FROM outbox_events
	`).Scan(&traceparent, &tracestate))
	require.Equal(t, "00-0102030405060708090a0b0c0d0e0f10-0102030405060708-01", traceparent)
	require.Equal(t, "vendor=value", tracestate)
}

func assertStoredEvent(t *testing.T, ctx context.Context, endpoint *postgresdb.Endpoint) {
	t.Helper()
	var idText, topic, aggregateType, aggregateID, aggregateKey, eventType string
	var payload []byte
	var routingHash int64
	var traceparent, tracestate sql.NullString
	err := endpoint.QueryRow(ctx, `
		SELECT id::text, topic, aggregatetype, aggregateid, aggregatekey, type,
		       payload, routing_hash, traceparent, tracestate
		FROM outbox_events
	`).Scan(
		&idText, &topic, &aggregateType, &aggregateID, &aggregateKey, &eventType,
		&payload, &routingHash, &traceparent, &tracestate,
	)
	require.NoError(t, err)
	id, err := uuid.Parse(idText)
	require.NoError(t, err)
	require.Equal(t, uuid.Version(7), id.Version())
	require.Equal(t, "orders", topic)
	require.Equal(t, "Order", aggregateType)
	require.Equal(t, "42", aggregateID)
	require.Equal(t, "Order::42", aggregateKey)
	require.Equal(t, "OrderPaid", eventType)
	require.Equal(t, []byte("wire"), payload)
	require.Equal(t, int64(xxh3.HashString("Order::42")), routingHash)
	require.False(t, traceparent.Valid)
	require.False(t, tracestate.Valid)
}

func openTestDatabase(t *testing.T, ctx context.Context) *postgresdb.DB {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx,
		"postgres:17.5-alpine",
		tcpostgres.WithDatabase("app"),
		tcpostgres.WithUsername("app"),
		tcpostgres.WithPassword("password"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(container)) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	db, err := postgresdb.Open(ctx, postgresdb.Config{Writer: postgresdb.EndpointConfig{DSN: dsn}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}

func applyMigrations(t *testing.T, ctx context.Context, endpoint *postgresdb.Endpoint, migrations fs.FS) {
	t.Helper()
	paths, err := fs.Glob(migrations, "*.up.sql")
	require.NoError(t, err)
	sort.Strings(paths)
	require.NotEmpty(t, paths)
	for _, path := range paths {
		migration, readErr := fs.ReadFile(migrations, path)
		require.NoError(t, readErr)
		_, execErr := endpoint.Exec(ctx, string(migration))
		require.NoError(t, execErr)
	}
}

func eventCount(t *testing.T, ctx context.Context, endpoint *postgresdb.Endpoint) int {
	t.Helper()
	var count int
	require.NoError(t, endpoint.QueryRow(ctx, `SELECT COUNT(*) FROM outbox_events`).Scan(&count))
	return count
}
