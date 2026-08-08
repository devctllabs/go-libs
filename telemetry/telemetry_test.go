package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

func TestOpenDisabledUsesNoopProvidersAndW3CPropagation(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "unsupported")
	t.Setenv("OTEL_METRICS_EXPORTER", "unsupported")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "unsupported")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "broken")

	runtime, err := Open(context.Background(), Config{})
	require.NoError(t, err)
	require.IsType(t, tracenoop.NewTracerProvider(), runtime.TracerProvider())
	require.IsType(t, metricnoop.NewMeterProvider(), runtime.MeterProvider())
	require.ElementsMatch(t, []string{"traceparent", "tracestate", "baggage"}, runtime.Propagator().Fields())

	member, err := baggage.NewMember("tenant", "acme")
	require.NoError(t, err)
	bag, err := baggage.New(member)
	require.NoError(t, err)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{2},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := baggage.ContextWithBaggage(trace.ContextWithSpanContext(context.Background(), spanContext), bag)
	carrier := propagation.MapCarrier{}
	runtime.Propagator().Inject(ctx, carrier)
	require.NotEmpty(t, carrier.Get("traceparent"))
	require.Equal(t, "tenant=acme", carrier.Get("baggage"))

	require.NoError(t, runtime.ForceFlush(context.Background()))
	require.NoError(t, runtime.Shutdown(context.Background()))
	require.NoError(t, runtime.Shutdown(context.Background()))
}

//nolint:paralleltest // The test temporarily replaces process-global OpenTelemetry providers.
func TestOpenDoesNotInstallGlobalProvidersOrPropagator(t *testing.T) {
	setSignalsNone(t)
	globalTracerProvider := tracenoop.NewTracerProvider()
	globalMeterProvider := metricnoop.NewMeterProvider()
	globalPropagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{})
	previousTracerProvider := otel.GetTracerProvider()
	previousMeterProvider := otel.GetMeterProvider()
	previousPropagator := otel.GetTextMapPropagator()
	otel.SetTracerProvider(globalTracerProvider)
	otel.SetMeterProvider(globalMeterProvider)
	otel.SetTextMapPropagator(globalPropagator)
	t.Cleanup(func() {
		otel.SetTracerProvider(previousTracerProvider)
		otel.SetMeterProvider(previousMeterProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})

	runtime, err := Open(context.Background(), validConfig())
	require.NoError(t, err)
	require.Equal(t, globalTracerProvider, otel.GetTracerProvider())
	require.Equal(t, globalMeterProvider, otel.GetMeterProvider())
	require.Equal(t, globalPropagator, otel.GetTextMapPropagator())
	require.NoError(t, runtime.Shutdown(context.Background()))
}

//nolint:paralleltest // setSignalsNone changes process environment for this test and its subtests.
func TestOpenEnabledValidatesServiceIdentity(t *testing.T) {
	setSignalsNone(t)

	tests := []struct {
		name   string
		config Config
		field  string
	}{
		{name: "service name", config: Config{Enabled: true, ServiceVersion: "1.2.3", DeploymentEnvironment: "test"}, field: "ServiceName"},
		{name: "service version", config: Config{Enabled: true, ServiceName: "orders", DeploymentEnvironment: "test"}, field: "ServiceVersion"},
		{name: "deployment environment", config: Config{Enabled: true, ServiceName: "orders", ServiceVersion: "1.2.3"}, field: "DeploymentEnvironment"},
		{name: "blank value", config: Config{Enabled: true, ServiceName: "orders", ServiceVersion: "   ", DeploymentEnvironment: "test"}, field: "ServiceVersion"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Open(context.Background(), tt.config)
			require.Nil(t, got)
			require.ErrorContains(t, err, tt.field)
		})
	}
}

func TestOpenEnabledAllowsSignalsToBeDisabledIndependently(t *testing.T) {
	setSignalsNone(t)
	t.Setenv("OTEL_PROPAGATORS", "b3")

	runtime, err := Open(context.Background(), validConfig())
	require.NoError(t, err)
	require.IsType(t, tracenoop.NewTracerProvider(), runtime.TracerProvider())
	require.IsType(t, metricnoop.NewMeterProvider(), runtime.MeterProvider())
	require.ElementsMatch(t, []string{"traceparent", "tracestate", "baggage"}, runtime.Propagator().Fields())
	require.NoError(t, runtime.Shutdown(context.Background()))
}

func TestOpenRejectsUnsupportedExporter(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "console")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")

	runtime, err := Open(context.Background(), validConfig())
	require.Nil(t, runtime)
	require.ErrorContains(t, err, "OTEL_TRACES_EXPORTER")
}

func TestOpenRejectsMultipleExporters(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp,none")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")

	runtime, err := Open(context.Background(), validConfig())
	require.Nil(t, runtime)
	require.ErrorContains(t, err, "OTEL_TRACES_EXPORTER")
}

func TestOpenRejectsUnsupportedProtocol(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http/json")

	runtime, err := Open(context.Background(), validConfig())
	require.Nil(t, runtime)
	require.ErrorContains(t, err, "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
}

func TestSignalProtocolOverridesGeneralProtocol(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "unsupported")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:4318")

	runtime, err := Open(context.Background(), validConfig())
	require.NoError(t, err)
	require.NoError(t, runtime.Shutdown(context.Background()))
}

func validConfig() Config {
	return Config{
		Enabled:               true,
		ServiceName:           "orders",
		ServiceVersion:        "1.2.3",
		DeploymentEnvironment: "test",
	}
}

func setSignalsNone(t *testing.T) {
	t.Helper()
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")
}
