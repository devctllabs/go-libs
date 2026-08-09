// Package telemetry initializes instance-owned OpenTelemetry trace and metric providers for Go
// services. Open reads standard OTLP exporter, trace SDK, metric reader, and resource environment
// variables while keeping provider and propagator installation explicit at the application
// composition root.
//
// The zero Config is disabled and performs no environment parsing, network setup, or background
// work. An enabled Runtime exports traces and metrics through OTLP by default; set
// OTEL_TRACES_EXPORTER or OTEL_METRICS_EXPORTER to none to disable a signal. Logs remain the
// application's responsibility. SetGlobalLogger only connects OpenTelemetry's own diagnostics to
// zap, and WithTraceContext adds trace correlation fields to application logs.
package telemetry
