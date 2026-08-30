//go:build integration

package kafkaoutbox_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/kafkaoutbox"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestCDCStoreAppendsIntoMaintainedUTCPartition(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.CDCMigrations())
	manager, err := kafkaoutbox.NewPartitionManager(db.Writer(), kafkaoutbox.PartitionManagerConfig{
		Granularity:     kafkaoutbox.PartitionDaily,
		AheadPartitions: 1,
		Retention:       48 * time.Hour,
	})
	require.NoError(t, err)
	require.NoError(t, manager.Maintain(ctx, time.Now()))
	store, err := kafkaoutbox.NewCDCStore(db.Writer())
	require.NoError(t, err)
	event := kafkaoutbox.Event[[]byte]{
		Topic: "orders", AggregateType: "Order", AggregateID: "42",
		Type: "OrderPaid", Value: []byte("wire"),
	}

	err = store.Append(ctx, event)
	require.ErrorIs(t, err, kafkaoutbox.ErrTransactionRequired)
	require.NoError(t, db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		return store.Append(txCtx, event)
	}))

	var topic, aggregateKey string
	var traceContext sql.NullString
	require.NoError(t, db.Writer().QueryRow(ctx, `
		SELECT topic, aggregatekey, tracingspancontext FROM outbox_events
	`).Scan(&topic, &aggregateKey, &traceContext))
	require.Equal(t, "orders", topic)
	require.Equal(t, "Order::42", aggregateKey)
	require.False(t, traceContext.Valid)
	var partitions int
	require.NoError(t, db.Writer().QueryRow(ctx, `
		SELECT COUNT(*)
		FROM pg_inherits
		WHERE inhparent = 'outbox_events'::regclass
	`).Scan(&partitions))
	require.Equal(t, 2, partitions)
}

func TestCDCStorePersistsDebeziumTracingSpanContext(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.CDCMigrations())
	manager, err := kafkaoutbox.NewPartitionManager(db.Writer(), kafkaoutbox.PartitionManagerConfig{
		Granularity: kafkaoutbox.PartitionDaily,
		Retention:   48 * time.Hour,
	})
	require.NoError(t, err)
	require.NoError(t, manager.Maintain(ctx, time.Now()))
	store, err := kafkaoutbox.NewCDCStore(db.Writer())
	require.NoError(t, err)
	traceState, err := trace.ParseTraceState("vendor=value")
	require.NoError(t, err)
	tracedCtx := trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SpanID:     trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
		TraceFlags: trace.FlagsSampled,
		TraceState: traceState,
	}))

	require.NoError(t, db.Writer().WithinTx(tracedCtx, func(txCtx context.Context) error {
		return store.Append(txCtx, kafkaoutbox.Event[[]byte]{
			Topic: "orders", AggregateType: "Order", AggregateID: "42", Type: "OrderPaid",
		})
	}))

	var traceContext string
	require.NoError(t, db.Writer().QueryRow(ctx, `
		SELECT tracingspancontext FROM outbox_events
	`).Scan(&traceContext))
	require.Equal(t,
		"traceparent=00-0102030405060708090a0b0c0d0e0f10-0102030405060708-01\ntracestate=vendor=value\n",
		traceContext,
	)
}

func TestPartitionManagerDropsOnlyFullyExpiredCanonicalPartitions(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.CDCMigrations())
	manager, err := kafkaoutbox.NewPartitionManager(db.Writer(), kafkaoutbox.PartitionManagerConfig{
		Granularity: kafkaoutbox.PartitionDaily,
		Retention:   48 * time.Hour,
	})
	require.NoError(t, err)
	require.NoError(t, manager.Maintain(ctx, time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)))
	require.NoError(t, manager.Maintain(ctx, time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)))

	rows, err := db.Writer().Query(ctx, `
		SELECT child.relname
		FROM pg_inherits
		JOIN pg_class child ON child.oid = inhrelid
		WHERE inhparent = 'outbox_events'::regclass
		ORDER BY child.relname
	`)
	require.NoError(t, err)
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []string{"outbox_events_p20260902"}, names)
}
