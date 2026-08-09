package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestStartRuntimeMetricsRegistersGoRuntimeInstruments(t *testing.T) {
	t.Parallel()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })

	require.NoError(t, startRuntimeMetrics(provider))
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))

	var names []string
	for _, scope := range data.ScopeMetrics {
		for _, metric := range scope.Metrics {
			names = append(names, metric.Name)
		}
	}
	require.Contains(t, names, "go.goroutine.count")
}
