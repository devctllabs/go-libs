package kafkaoutbox

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestOTelObserverRecordsOutboxMetricsAndLinkedPublishSpan(t *testing.T) {
	t.Parallel()
	reader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, meterProvider.Shutdown(context.Background())) })
	recorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, tracerProvider.Shutdown(context.Background())) })
	observer, err := NewOTelObserver(OTelObserverConfig{
		MeterProvider: meterProvider, TracerProvider: tracerProvider, MaxBatchSpanLinks: 1,
	})
	require.NoError(t, err)
	first := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}, Remote: true,
	})
	second := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{2}, SpanID: trace.SpanID{2}, Remote: true,
	})
	ctx, done := observer.StartOperation(context.Background(), Operation{
		Phase: OperationPublish, Generation: 1, ShardID: 2,
		Records: []RecordMetadata{
			{Topic: "orders", Context: trace.ContextWithSpanContext(context.Background(), first)},
			{Topic: "orders", Context: trace.ContextWithSpanContext(context.Background(), second)},
		},
	})
	done(nil)
	observer.Enqueued(ctx)
	observer.BatchCompleted(ctx, BatchResult{Outcome: BatchDelivered, Size: 2})
	observer.Retry(ctx, RetryEvent{Attempt: 1, NextDelay: time.Second, Err: errors.New("secret")})
	observer.FencingConflict(ctx, FencingEvent{Phase: OperationFinalize})
	observer.TopologyReconciled(ctx, TopologyResult{Changed: true, ShardCount: 4})

	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	metrics := outboxMetricsByName(data)
	for _, name := range []string{
		"kafkaoutbox.operation.attempts",
		"kafkaoutbox.operation.duration",
		"kafkaoutbox.operation.inflight",
		"kafkaoutbox.enqueue.events",
		"kafkaoutbox.worker.batches",
		"kafkaoutbox.worker.events",
		"kafkaoutbox.worker.batch.size",
		"kafkaoutbox.worker.retries",
		"kafkaoutbox.worker.fencing_conflicts",
		"kafkaoutbox.topology.changes",
		"kafkaoutbox.topology.shards",
	} {
		require.Contains(t, metrics, name)
	}
	_, ok := metrics["kafkaoutbox.topology.shards"].Data.(metricdata.Gauge[int64])
	require.True(t, ok)
	require.NotContains(t, fmt.Sprint(data), "secret")

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "kafkaoutbox publish", spans[0].Name())
	require.Equal(t, trace.SpanKindProducer, spans[0].SpanKind())
	require.Len(t, spans[0].Links(), 1)
	require.Equal(t, first, spans[0].Links()[0].SpanContext)
}

func TestNewOTelObserverRejectsNegativeLinkCap(t *testing.T) {
	t.Parallel()
	_, err := NewOTelObserver(OTelObserverConfig{MaxBatchSpanLinks: -1})
	require.Error(t, err)
}

func outboxMetricsByName(data metricdata.ResourceMetrics) map[string]metricdata.Metrics {
	metrics := make(map[string]metricdata.Metrics)
	for _, scope := range data.ScopeMetrics {
		for _, value := range scope.Metrics {
			metrics[value.Name] = value
		}
	}
	return metrics
}
