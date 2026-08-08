// Package health coordinates transport-neutral liveness and readiness probes.
//
// Liveness is always process-local and never invokes component checks. Readiness runs explicitly
// registered critical and non-critical checks concurrently under one deadline. Critical failures
// make readiness fail; non-critical failures remain visible in reports and observers without
// removing the instance from traffic.
//
// Probes are immutable after New and start workers only for an active Readiness call. Checkers must
// honor context cancellation; a checker that ignores it can continue running after Readiness has
// returned. Observers may be called concurrently by separate Readiness calls and must return
// promptly. Raw checker errors are available only to observers, never to transport-safe reports.
package health
