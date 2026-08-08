package txmanager

import (
	"context"
	"errors"
)

// Backend adapts one database driver's native transaction type to Coordinator.
//
//go:generate go tool mockgen -destination mocks/backend.gen.go -package mocks . Backend
type Backend[T any] interface {
	// Begin starts a native transaction using spec.
	Begin(ctx context.Context, spec BeginSpec) (T, error)
	// Commit commits tx.
	Commit(ctx context.Context, tx T) error
	// Rollback rolls back tx.
	Rollback(ctx context.Context, tx T) error
}

// BeginSpec describes the transaction requested from a Backend.
type BeginSpec struct {
	Role      Role
	Isolation *Isolation
}

// Coordinator coordinates transactions for one database instance.
type Coordinator[T any] struct {
	backend Backend[T]
	key     *contextKey
	reader  Manager
	writer  Manager
}

type contextKey struct{}

type transactionState[T any] struct {
	tx        T
	role      Role
	isolation *Isolation
}

// NewCoordinator constructs a transaction coordinator for backend.
func NewCoordinator[T any](backend Backend[T]) (*Coordinator[T], error) {
	if backend == nil {
		return nil, errors.New("txmanager: backend must not be nil")
	}
	coordinator := &Coordinator[T]{backend: backend, key: &contextKey{}}
	coordinator.reader = &boundManager[T]{coordinator: coordinator, role: RoleReader}
	coordinator.writer = &boundManager[T]{coordinator: coordinator, role: RoleWriter}
	return coordinator, nil
}

// Reader returns the read-only transaction manager.
func (c *Coordinator[T]) Reader() Manager {
	return c.reader
}

// Writer returns the read-write transaction manager.
func (c *Coordinator[T]) Writer() Manager {
	return c.writer
}

// Current returns the native transaction associated with ctx for this coordinator.
func (c *Coordinator[T]) Current(ctx context.Context) (T, bool) {
	var zero T
	if ctx == nil {
		return zero, false
	}
	state, ok := ctx.Value(c.key).(transactionState[T])
	if !ok {
		return zero, false
	}
	return state.tx, true
}

func (c *Coordinator[T]) state(ctx context.Context) (transactionState[T], bool) {
	if ctx == nil {
		return transactionState[T]{}, false
	}
	state, ok := ctx.Value(c.key).(transactionState[T])
	return state, ok
}
