package oidcsession

import (
	"context"
	"time"
)

// Operation identifies a bounded OIDC/session operation for diagnostics.
type Operation string

const (
	OperationDiscovery Operation = "discovery"
	OperationLogin     Operation = "login"
	OperationCallback  Operation = "callback"
	OperationRefresh   Operation = "refresh"
	OperationLogout    Operation = "logout"
	OperationSession   Operation = "session"
)

// Observation contains operational diagnostics. Callers must not attach credentials or claims.
type Observation struct {
	Operation Operation
	Err       error
	Duration  time.Duration
	Retry     bool
}

// Observer receives operational failures and retries.
type Observer interface {
	// Observe records observation promptly and must not retain request credentials.
	Observe(ctx context.Context, observation Observation)
}

// ObserverFunc adapts a function to Observer.
type ObserverFunc func(ctx context.Context, observation Observation)

// Observe implements Observer.
func (function ObserverFunc) Observe(ctx context.Context, observation Observation) {
	function(ctx, observation)
}

func observe(ctx context.Context, observer Observer, observation Observation) {
	if observer != nil {
		observer.Observe(ctx, observation)
	}
}
