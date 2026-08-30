package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestOTelObserverCreatesCappedDeduplicatedAttemptLinks(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	observer, err := NewOTelObserver(OTelObserverConfig{
		TracerProvider:    provider,
		MaxBatchSpanLinks: 1,
	})
	require.NoError(t, err)
	first := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}, Remote: true,
	})
	second := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{2}, SpanID: trace.SpanID{2}, Remote: true,
	})

	_, done := observer.StartAttempt(context.Background(), Attempt{
		Phase:   AttemptHandler,
		Attempt: 1,
		Records: []RecordMetadata{
			{Topic: "invoices", Partition: 1, Context: trace.ContextWithSpanContext(context.Background(), first)},
			{Topic: "invoices", Partition: 1, Context: trace.ContextWithSpanContext(context.Background(), first)},
			{Topic: "invoices", Partition: 1, Context: trace.ContextWithSpanContext(context.Background(), second)},
		},
	})
	done(nil)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "kafka handler attempt", spans[0].Name())
	require.Len(t, spans[0].Links(), 1)
	require.Equal(t, first, spans[0].Links()[0].SpanContext)
}

func TestOTelObserverSuppliesStandardFranzHooks(t *testing.T) {
	t.Parallel()

	observer, err := NewOTelObserver(OTelObserverConfig{})
	require.NoError(t, err)

	require.Len(t, observerHooks(observer, "invoice-consumer"), 2)
	require.Empty(t, observerHooks(noopObserver{}, "invoice-consumer"))
}
