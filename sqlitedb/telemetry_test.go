package sqlitedb_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devctllabs/go-libs/sqlitedb"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

//nolint:paralleltest // The test temporarily replaces the process-global tracer provider.
func TestOpenUsesExplicitTracerAndProtectsQueryData(t *testing.T) {
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
	db, err := sqlitedb.Open(context.Background(), sqlitedb.Config{
		DSN: filepath.Join(t.TempDir(), "telemetry.sqlite"),
		Telemetry: sqlitedb.Telemetry{
			TracerProvider: provider,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	var value string
	err = db.Writer().QueryRowContext(context.Background(), `SELECT ?`, "sensitive-argument").Scan(&value)
	require.NoError(t, err)
	require.Equal(t, "sensitive-argument", value)

	serialized := serializeSpans(recorder.Ended())
	require.NotEmpty(t, recorder.Ended())
	require.Empty(t, globalRecorder.Ended())
	require.NotContains(t, serialized, "sensitive-argument")
	require.NotContains(t, serialized, "SELECT ?")
}

func TestIncludeQueryTextRecordsStatementButNotArguments(t *testing.T) {
	t.Parallel()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	db, err := sqlitedb.Open(context.Background(), sqlitedb.Config{
		DSN: filepath.Join(t.TempDir(), "query-text.sqlite"),
		Telemetry: sqlitedb.Telemetry{
			TracerProvider:   provider,
			IncludeQueryText: true,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	var value string
	err = db.Writer().QueryRowContext(
		context.Background(),
		`SELECT ? AS visible_statement`,
		"sensitive-argument",
	).Scan(&value)
	require.NoError(t, err)

	serialized := serializeSpans(recorder.Ended())
	require.Contains(t, serialized, "SELECT ? AS visible_statement")
	require.NotContains(t, serialized, "sensitive-argument")
}

func serializeSpans(spans []sdktrace.ReadOnlySpan) string {
	var serialized strings.Builder
	for _, span := range spans {
		_, _ = fmt.Fprintln(&serialized, span.Name())
		for _, attribute := range span.Attributes() {
			_, _ = fmt.Fprintf(&serialized, "%s=%v\n", attribute.Key, attribute.Value.AsInterface())
		}
	}
	return serialized.String()
}
