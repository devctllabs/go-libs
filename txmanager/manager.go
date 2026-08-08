package txmanager

import (
	"context"
	"errors"
	"fmt"
)

// Role identifies the database endpoint that starts a transaction.
type Role uint8

const (
	// RoleReader starts a read-only transaction.
	RoleReader Role = iota + 1
	// RoleWriter starts a read-write transaction.
	RoleWriter
)

// Manager executes callbacks inside transactions owned by one endpoint role.
type Manager interface {
	// WithinTx executes fn in a transaction and commits only when fn returns nil.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error, options ...Option) error
}

// Option configures one WithinTx invocation.
type Option interface {
	apply(options *txOptions) error
}

type txOptions struct {
	isolation *Isolation
}

type boundManager[T any] struct {
	coordinator *Coordinator[T]
	role        Role
}

func (m *boundManager[T]) WithinTx(
	ctx context.Context,
	fn func(ctx context.Context) error,
	options ...Option,
) error {
	if ctx == nil {
		return errors.New("txmanager: context must not be nil")
	}
	if fn == nil {
		return errors.New("txmanager: callback must not be nil")
	}
	configured, err := applyOptions(options)
	if err != nil {
		return err
	}
	if active, ok := m.coordinator.state(ctx); ok {
		if active.role == RoleReader && m.role == RoleWriter {
			return ErrReadOnlyEscalation
		}
		if configured.isolation != nil && !sameIsolation(active.isolation, configured.isolation) {
			return ErrIsolationMismatch
		}
		return fn(ctx)
	}

	spec := BeginSpec{Role: m.role, Isolation: configured.isolation}
	tx, err := m.coordinator.backend.Begin(ctx, spec)
	if err != nil {
		return fmt.Errorf("txmanager: begin transaction: %w", err)
	}
	txCtx := context.WithValue(ctx, m.coordinator.key, transactionState[T]{
		tx:        tx,
		role:      m.role,
		isolation: configured.isolation,
	})
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = m.coordinator.backend.Rollback(ctx, tx)
			panic(recovered)
		}
	}()
	if callbackErr := fn(txCtx); callbackErr != nil {
		rollbackErr := m.coordinator.backend.Rollback(ctx, tx)
		return errors.Join(callbackErr, wrapRollback(rollbackErr))
	}
	if err := m.coordinator.backend.Commit(ctx, tx); err != nil {
		return fmt.Errorf("txmanager: commit transaction: %w", err)
	}
	return nil
}

func applyOptions(options []Option) (txOptions, error) {
	configured := txOptions{}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option.apply(&configured); err != nil {
			return txOptions{}, err
		}
	}
	return configured, nil
}

func sameIsolation(left *Isolation, right *Isolation) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func wrapRollback(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("txmanager: rollback transaction: %w", err)
}
