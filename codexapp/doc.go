// Package codexapp provides a process-owning Go client for the Codex App Server JSONL JSON-RPC
// protocol.
//
// Open starts exactly one `codex app-server --listen stdio://` process. Client.Close owns its
// shutdown. Supervisor is optional and restarts process sessions, but deliberately never replays
// a protocol request because a written request may already have taken effect.
//
// Generated protocol DTOs and the selected schema snapshot stay internal. Callers use the stable
// projections, typed inputs, turn handles, event streams, and server-request handler in this
// package. OpenTelemetry providers are supplied explicitly; the package does not read or mutate
// OpenTelemetry globals.
package codexapp
