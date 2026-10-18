// Package dbmigrate runs a recursive selection of SQL migrations with golang-migrate.
package dbmigrate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sync"

	"github.com/golang-migrate/migrate/v4"
)

var (
	// ErrInvalidConfig identifies a missing or invalid construction argument.
	ErrInvalidConfig = errors.New("invalid migration configuration")
	// ErrInvalidSteps identifies a step count that is not positive.
	ErrInvalidSteps = errors.New("migration steps must be positive")
	// ErrClosed identifies an operation on a closed Migrator.
	ErrClosed = errors.New("migrator is closed")
)

// Config selects a migration database and directories relative to the supplied filesystem.
type Config struct {
	// DatabaseURL uses the URL scheme and query parameters of a registered
	// golang-migrate database adapter. The caller imports the selected adapter.
	DatabaseURL string
	// Directories is a nonempty list of relative fs.ValidPath directories.
	// Use "." to select the entire filesystem. Order does not affect execution.
	Directories []string
}

// Migrator owns a migration database connection and must be constructed with New.
// Concurrent operations return migrate.ErrLocked; Close waits for the active call.
// Filesystem ownership remains with the caller.
type Migrator struct {
	mu       sync.Mutex
	engine   *migrate.Migrate
	closed   bool
	closeErr error
}

// New recursively indexes SQL files in the selected directories, then opens the
// migration database. Opening may create the driver's version-tracking table,
// but does not apply migrations. The index is fixed at construction; SQL bodies
// are read when the engine requests them. The caller must keep fsys available
// until Close and must not modify migrations during an operation.
//
// Invalid arguments return ErrInvalidConfig; filesystem and upstream errors
// remain available through errors.Is and errors.As. Duplicate version/direction
// pairs are rejected, but missing up/down counterparts and invalid filenames
// follow the upstream source semantics. No environment directory is implicit.
func New(fsys fs.FS, cfg Config) (*Migrator, error) {
	if fsys == nil || len(cfg.Directories) == 0 || cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("%w: filesystem, directories and database URL are required", ErrInvalidConfig)
	}
	for _, directory := range cfg.Directories {
		if !fs.ValidPath(directory) {
			return nil, fmt.Errorf("%w: invalid directory %q", ErrInvalidConfig, directory)
		}
	}
	src, err := newSource(fsys, cfg.Directories)
	if err != nil {
		return nil, err
	}
	engine, err := migrate.NewWithSourceInstance("recursive", src, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return &Migrator{engine: engine}, nil
}

// Up applies all pending migrations in increasing numeric version order.
// A canceled ctx prevents execution; cancellation after execution starts does
// not interrupt it. The call waits for the complete upstream operation and
// preserves its result, including migrate.ErrNoChange.
func (m *Migrator) Up(ctx context.Context) error {
	return m.run(ctx, (*migrate.Migrate).Up)
}

// UpN applies n migrations upward, requiring n > 0 or returning ErrInvalidSteps.
// Like upstream Steps, it may apply available migrations before returning an
// error when fewer than n remain. Context behavior matches Up.
func (m *Migrator) UpN(ctx context.Context, n int) error {
	if n <= 0 {
		return ErrInvalidSteps
	}
	return m.run(ctx, func(engine *migrate.Migrate) error { return engine.Steps(n) })
}

// Down applies n migrations in decreasing numeric version order, requiring
// n > 0 or returning ErrInvalidSteps. Like UpN, a shortage can return an error
// after partial execution. Context behavior matches Up.
func (m *Migrator) Down(ctx context.Context, n int) error {
	if n <= 0 {
		return ErrInvalidSteps
	}
	return m.run(ctx, func(engine *migrate.Migrate) error { return engine.Steps(-n) })
}

// Force sets a version >= -1 and clears dirty without executing SQL or checking
// that the version exists in the source. Version -1 means no applied migration.
// The caller must repair the database before using Force for recovery.
func (m *Migrator) Force(version int) error {
	if err := m.acquire(); err != nil {
		return err
	}
	defer m.mu.Unlock()
	return m.engine.Force(version)
}

// Version returns the current migration version and dirty state.
// An unmigrated database returns migrate.ErrNilVersion.
func (m *Migrator) Version() (uint, bool, error) {
	if err := m.acquire(); err != nil {
		return 0, false, err
	}
	defer m.mu.Unlock()
	return m.engine.Version()
}

// Close waits for the active operation and closes the migration database once.
// It never closes the caller's filesystem. Repeated calls return the same error;
// subsequent operations return ErrClosed.
func (m *Migrator) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		m.closed = true
		sourceErr, databaseErr := m.engine.Close()
		m.closeErr = errors.Join(sourceErr, databaseErr)
	}
	return m.closeErr
}

func (m *Migrator) acquire() error {
	if !m.mu.TryLock() {
		return migrate.ErrLocked
	}
	if m.closed {
		m.mu.Unlock()
		return ErrClosed
	}
	return nil
}

func (m *Migrator) run(ctx context.Context, operation func(*migrate.Migrate) error) error {
	if err := m.acquire(); err != nil {
		return err
	}
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	return operation(m.engine)
}
