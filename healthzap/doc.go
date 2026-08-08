// Package healthzap logs health check failures and recoveries with a caller-owned zap logger.
// Every critical failure is logged at error, every non-critical failure at warn, and the first
// success after failures at info. Ordinary successes are silent. Callers own logger construction,
// redaction, output, and lifecycle.
package healthzap
