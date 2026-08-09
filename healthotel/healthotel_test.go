package healthotel_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/health"
	"github.com/devctllabs/go-libs/healthotel"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestObserverRecordsStatusExecutionsAndDurationWithoutRawErrors(t *testing.T) {
	t.Parallel()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	observer, err := healthotel.New(provider)
	require.NoError(t, err)

	observer.Observe(context.Background(), health.Observation{
		Name:     "postgres",
		Critical: true,
		Outcome:  health.OutcomeError,
		Duration: 125 * time.Millisecond,
		Err:      errors.New("postgres://secret@db.internal unavailable"),
	})
	observer.Observe(context.Background(), health.Observation{
		Name:     "postgres",
		Critical: true,
		Outcome:  health.OutcomeOK,
		Duration: 25 * time.Millisecond,
	})

	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	metrics := metricsByName(data)
	require.Contains(t, metrics, "health.check.status")
	require.Contains(t, metrics, "health.check.executions")
	require.Contains(t, metrics, "health.check.duration")
	require.NotContains(t, fmt.Sprint(data), "secret")

	status, ok := metrics["health.check.status"].Data.(metricdata.Gauge[int64])
	require.True(t, ok)
	require.Len(t, status.DataPoints, 1)
	require.Equal(t, int64(1), status.DataPoints[0].Value)
	name, ok := status.DataPoints[0].Attributes.Value(attribute.Key("check.name"))
	require.True(t, ok)
	require.Equal(t, "postgres", name.AsString())
	critical, ok := status.DataPoints[0].Attributes.Value(attribute.Key("check.critical"))
	require.True(t, ok)
	require.True(t, critical.AsBool())

	executions, ok := metrics["health.check.executions"].Data.(metricdata.Sum[int64])
	require.True(t, ok)
	require.Len(t, executions.DataPoints, 2)
	var executionCount int64
	for _, point := range executions.DataPoints {
		executionCount += point.Value
		result, found := point.Attributes.Value(attribute.Key("result"))
		require.True(t, found)
		require.Contains(t, []string{"ok", "error"}, result.AsString())
	}
	require.Equal(t, int64(2), executionCount)

	duration, ok := metrics["health.check.duration"].Data.(metricdata.Histogram[float64])
	require.True(t, ok)
	require.Len(t, duration.DataPoints, 2)
	var durationCount uint64
	for _, point := range duration.DataPoints {
		durationCount += point.Count
	}
	require.Equal(t, uint64(2), durationCount)
}

func TestNewRejectsNilMeterProvider(t *testing.T) {
	t.Parallel()
	_, err := healthotel.New(nil)
	require.ErrorContains(t, err, "must not be nil")
}

func TestObserverNormalizesUnknownOutcomeToError(t *testing.T) {
	t.Parallel()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	observer, err := healthotel.New(provider)
	require.NoError(t, err)

	observer.Observe(context.Background(), health.Observation{
		Name: "postgres", Outcome: health.Outcome("unbounded-value"),
	})
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	require.NotContains(t, fmt.Sprint(data), "unbounded-value")

	executions := metricsByName(data)["health.check.executions"].Data.(metricdata.Sum[int64])
	require.Len(t, executions.DataPoints, 1)
	result, ok := executions.DataPoints[0].Attributes.Value(attribute.Key("result"))
	require.True(t, ok)
	require.Equal(t, "error", result.AsString())
}

func metricsByName(data metricdata.ResourceMetrics) map[string]metricdata.Metrics {
	metrics := make(map[string]metricdata.Metrics)
	for _, scope := range data.ScopeMetrics {
		for _, value := range scope.Metrics {
			metrics[value.Name] = value
		}
	}
	return metrics
}
