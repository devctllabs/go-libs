// Package lifecycle coordinates long-running application tasks and one common graceful shutdown.
//
// The caller owns signal handling, dependency construction, and configuration. Tasks must either
// honor their context or be stopped by Config.Shutdown. Run calls shutdown before waiting for all
// tasks, so servers whose Serve method returns only after Shutdown are supported without extra
// goroutines in application code. One-shot commands can use Shutdown directly to run the same
// fresh, bounded cleanup convention without starting lifecycle tasks.
package lifecycle
