// Package healthotel records health check observations with an explicitly supplied OpenTelemetry
// MeterProvider. It creates health.check.status, health.check.executions, and
// health.check.duration instruments once in New. Check names must be a bounded configured set;
// raw checker errors are never used as metric attributes.
package healthotel
