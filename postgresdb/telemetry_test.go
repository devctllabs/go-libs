package postgresdb

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

//nolint:paralleltest // The test temporarily replaces the process-global tracer provider.
func TestOpenConfiguresExplicitQueryTracingWithoutConnecting(t *testing.T) {
	globalRecorder := tracetest.NewSpanRecorder()
	globalProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(globalRecorder))
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(globalProvider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		require.NoError(t, globalProvider.Shutdown(context.Background()))
	})

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	db, err := Open(context.Background(), Config{
		Writer: EndpointConfig{DSN: "postgres://user:password@127.0.0.1:1/app"},
		Telemetry: Telemetry{
			TracerProvider: provider,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	tracer := db.writerPool.Config().ConnConfig.Tracer
	parentCtx, parentSpan := provider.Tracer("postgresdb-test").Start(context.Background(), "parent")
	traceCtx := tracer.TraceQueryStart(parentCtx, nil, pgx.TraceQueryStartData{
		SQL:  `SELECT $1::text AS hidden_statement`,
		Args: []any{"sensitive-argument"},
	})
	tracer.TraceQueryEnd(traceCtx, nil, pgx.TraceQueryEndData{})
	parentSpan.End()

	serialized := serializeUnitSpans(recorder.Ended())
	require.NotEmpty(t, recorder.Ended())
	require.Empty(t, globalRecorder.Ended())
	require.Contains(t, serialized, "db.role=writer")
	require.NotContains(t, serialized, "hidden_statement")
	require.NotContains(t, serialized, "sensitive-argument")
}

func TestIncludeQueryTextRecordsStatementButNeverParameters(t *testing.T) {
	t.Parallel()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	db, err := Open(context.Background(), Config{
		Writer: EndpointConfig{DSN: "postgres://user:password@127.0.0.1:1/app"},
		Telemetry: Telemetry{
			TracerProvider:   provider,
			IncludeQueryText: true,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	tracer := db.writerPool.Config().ConnConfig.Tracer
	parentCtx, parentSpan := provider.Tracer("postgresdb-test").Start(context.Background(), "parent")
	traceCtx := tracer.TraceQueryStart(parentCtx, nil, pgx.TraceQueryStartData{
		SQL:  `SELECT $1::text AS visible_statement`,
		Args: []any{"sensitive-argument"},
	})
	tracer.TraceQueryEnd(traceCtx, nil, pgx.TraceQueryEndData{})
	parentSpan.End()

	serialized := serializeUnitSpans(recorder.Ended())
	require.Contains(t, serialized, "SELECT $1::text AS visible_statement")
	require.NotContains(t, serialized, "sensitive-argument")
}

func serializeUnitSpans(spans []sdktrace.ReadOnlySpan) string {
	var serialized strings.Builder
	for _, span := range spans {
		_, _ = fmt.Fprintln(&serialized, span.Name())
		for _, attribute := range span.Attributes() {
			_, _ = fmt.Fprintf(&serialized, "%s=%v\n", attribute.Key, attribute.Value.AsInterface())
		}
	}
	return serialized.String()
}
