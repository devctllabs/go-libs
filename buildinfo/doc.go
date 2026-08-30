// Package buildinfo reads build metadata embedded in the current Go executable.
//
// The package reports the main module version, source revision and Go toolchain version without
// reading the environment, filesystem, network or version control system. Applications decide how
// to attach the returned values to logs, telemetry, error reports or management endpoints.
package buildinfo
