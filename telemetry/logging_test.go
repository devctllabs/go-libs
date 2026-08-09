package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestSetGlobalLoggerRejectsNil(t *testing.T) {
	t.Parallel()
	require.Error(t, SetGlobalLogger(nil))
}

//nolint:paralleltest // SetGlobalLogger intentionally mutates process-global OpenTelemetry state.
func TestSetGlobalLoggerAcceptsZapLogger(t *testing.T) {
	require.NoError(t, SetGlobalLogger(zap.NewNop()))
}

func TestWithTraceContextReturnsSameLoggerWithoutValidSpan(t *testing.T) {
	t.Parallel()
	logger := zap.NewNop()
	require.Same(t, logger, WithTraceContext(context.Background(), logger))
}

func TestWithTraceContextAddsTraceAndSpanIDs(t *testing.T) {
	t.Parallel()
	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1, 2, 3},
		SpanID:  trace.SpanID{4, 5, 6},
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

	WithTraceContext(ctx, logger).Info("handled")

	entries := observed.AllUntimed()
	require.Len(t, entries, 1)
	entry := entries[0]
	require.Equal(t, spanContext.TraceID().String(), entry.ContextMap()["trace_id"])
	require.Equal(t, spanContext.SpanID().String(), entry.ContextMap()["span_id"])
}
