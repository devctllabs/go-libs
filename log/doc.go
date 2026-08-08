// Package log constructs JSON-encoded zap loggers for application diagnostics.
//
// New always uses zap's production JSON encoder with ISO 8601 timestamps. Log
// entries go to stderr by default, and enabling stacktraces adds them only to
// Error-level and higher entries. The package does not provide a plain-text
// mode or install a global zap logger.
//
// # Integration
//
// Pass the returned *zap.Logger explicitly through constructors. Do not wrap it
// in a custom logger interface solely for dependency injection, and do not use
// zap.L, zap.S, or zap.ReplaceGlobals. Add application-wide fields once when
// constructing the logger and use Named for static component identity.
//
// The caller owns the logger lifecycle. Call Sync when required by the selected
// output and handle output-specific Sync errors at the application boundary.
// WithOutput replaces stderr for tests or integrations; it does not close the
// supplied writer.
//
// # Testing
//
// Prefer zap.NewNop for ordinary unit tests, zaptest.NewLogger when logs should
// be attached to testing.T, and zaptest/observer when a test asserts structured
// entries. Use WithOutput when testing this package's concrete JSON encoding or
// an application logger factory.
package log
