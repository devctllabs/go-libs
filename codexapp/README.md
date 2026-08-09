# codexapp

`codexapp` embeds Codex App Server as a process-owned, bidirectional JSON-RPC dependency. It is
not an HTTP/OpenAPI client: App Server speaks newline-delimited JSON-RPC over stdio and can send
notifications and approval/input requests back to the host.

The package intentionally separates three compatibility surfaces:

- selected Codex JSON schemas are checked in under `internal/protocol/schema`;
- quicktype output is internal and records the generating Codex CLI version;
- the public package exposes small stable projections instead of generated DTOs.

## Lifecycle

`Open` starts one process and `Client.Close` stops it. For long-running applications,
`Supervisor.Run` can restart an exited process and `Supervisor.Client` waits for a ready
generation. A restart creates a new session: threads must be resumed explicitly and failed calls
are never replayed.

Turn notifications are delivered through a bounded per-turn `EventStream`. A slow consumer gets
`ErrEventOverflow`; it cannot block the protocol reader or unrelated RPC responses. A process exit
finishes active handles with `SessionLostError`.

Server-initiated approval, user-input, permission, and MCP elicitation requests run outside the
reader loop. With no handler, or when a handler fails, the response is fail-closed.

## Permissions and telemetry

Beta permission profiles are opt-in through `RequiredCapabilities`. Root paths must be absolute,
filesystem roots are rejected, and `thread/start` receives a deterministic deny-by-default
profile. There is no silent fallback when the capability probe fails.

Pass explicit OpenTelemetry tracer/meter providers and a propagator through `Telemetry`. Nil
providers are no-op. The package never installs global providers or propagators.

## Updating the protocol snapshot

The repository pins Go, Node, Codex CLI, and quicktype in the root `mise.toml`/`mise.lock`.

```sh
mise run codexapp:generate
mise run codexapp:check-generated
mise run codexapp:test
```

Generated files must be refreshed deliberately with the pinned tools and reviewed together with
public mapping changes.
