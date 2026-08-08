package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func openTraceExporter(ctx context.Context, cfg signalConfig) (sdktrace.SpanExporter, error) {
	if cfg.exporter == exporterNone {
		return nil, nil
	}
	var (
		exporter sdktrace.SpanExporter
		err      error
	)
	switch cfg.protocol {
	case protocolHTTPProtobuf:
		exporter, err = otlptracehttp.New(ctx)
	case protocolGRPC:
		exporter, err = otlptracegrpc.New(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("open OTLP trace exporter: %w", err)
	}
	return exporter, nil
}

func openMetricExporter(ctx context.Context, cfg signalConfig) (sdkmetric.Exporter, error) {
	if cfg.exporter == exporterNone {
		return nil, nil
	}
	var (
		exporter sdkmetric.Exporter
		err      error
	)
	switch cfg.protocol {
	case protocolHTTPProtobuf:
		exporter, err = otlpmetrichttp.New(ctx)
	case protocolGRPC:
		exporter, err = otlpmetricgrpc.New(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("open OTLP metric exporter: %w", err)
	}
	return exporter, nil
}

func shutdownTraceExporter(ctx context.Context, exporter sdktrace.SpanExporter) error {
	if exporter == nil {
		return nil
	}
	return exporter.Shutdown(ctx)
}
