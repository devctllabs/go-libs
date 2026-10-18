package dbmigrate_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/devctllabs/go-libs/dbmigrate"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/stub"
	"github.com/stretchr/testify/require"
)

func TestMigratorCanceledBeforeExecution(t *testing.T) {
	t.Parallel()
	m, _ := newSQLiteMigrator(t, stepMigrations(), ".")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, m.Up(ctx), context.Canceled)
	require.ErrorIs(t, m.UpN(ctx, 1), context.Canceled)
	require.ErrorIs(t, m.Down(ctx, 1), context.Canceled)
	_, _, err := m.Version()
	require.ErrorIs(t, err, migrate.ErrNilVersion)
}

func TestMigratorCancellationAfterStartWaitsForWholeOperation(t *testing.T) {
	t.Parallel()
	for _, direction := range []string{"up", "upN", "down"} {
		t.Run(direction, func(t *testing.T) {
			t.Parallel()
			m, driver, release := newControlledMigrator(t, 32)
			operation := m.Up
			if direction == "upN" {
				operation = func(ctx context.Context) error { return m.UpN(ctx, 32) }
			}
			if direction == "down" {
				driver.CurrentVersion = 32
				operation = func(ctx context.Context) error { return m.Down(ctx, 32) }
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			result := make(chan error, 1)
			go func() { result <- operation(ctx) }()
			waitStarted(t, driver.started)
			cancel()
			assertStillRunning(t, result)
			release()
			require.NoError(t, waitResult(t, result))
			require.Len(t, driver.MigrationSequence, 32)
			if direction == "down" {
				_, _, err := m.Version()
				require.ErrorIs(t, err, migrate.ErrNilVersion)
				require.NoError(t, m.UpN(t.Context(), 1))
				assertVersion(t, m, 1, false)
			} else {
				assertVersion(t, m, 32, false)
				require.NoError(t, m.Down(t.Context(), 1))
				assertVersion(t, m, 31, false)
			}
		})
	}
}

func TestMigratorCancellationAfterStartPreservesExecutionError(t *testing.T) {
	t.Parallel()
	m, driver, release := newControlledMigrator(t, 1)
	driver.runErr = errors.New("database failure")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- m.Up(ctx) }()
	waitStarted(t, driver.started)
	cancel()
	release()
	err := waitResult(t, result)
	require.NotErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, err, driver.runErr)
}

func TestMigratorConcurrentOperationsAndClose(t *testing.T) {
	t.Parallel()
	m, driver, release := newControlledMigrator(t, 1)
	result := make(chan error, 1)
	go func() { result <- m.Up(t.Context()) }()
	waitStarted(t, driver.started)
	require.ErrorIs(t, m.Up(t.Context()), migrate.ErrLocked)
	require.ErrorIs(t, m.UpN(t.Context(), 1), migrate.ErrLocked)
	require.ErrorIs(t, m.Down(t.Context(), 1), migrate.ErrLocked)
	require.ErrorIs(t, m.Force(0), migrate.ErrLocked)
	_, _, err := m.Version()
	require.ErrorIs(t, err, migrate.ErrLocked)
	closed := make(chan error, 1)
	go func() { closed <- m.Close() }()
	assertStillRunning(t, closed)
	require.Zero(t, driver.closes.Load())
	release()
	require.NoError(t, waitResult(t, result))
	require.NoError(t, waitResult(t, closed))
	require.NoError(t, m.Close())
	require.Equal(t, int32(1), driver.closes.Load())
	require.ErrorIs(t, m.Up(t.Context()), dbmigrate.ErrClosed)
	require.ErrorIs(t, m.UpN(t.Context(), 1), dbmigrate.ErrClosed)
	require.ErrorIs(t, m.Down(t.Context(), 1), dbmigrate.ErrClosed)
	require.ErrorIs(t, m.Force(0), dbmigrate.ErrClosed)
	_, _, err = m.Version()
	require.ErrorIs(t, err, dbmigrate.ErrClosed)
}

func TestMigratorClosePreservesErrorsAndBorrowedFilesystem(t *testing.T) {
	t.Parallel()
	m, driver, _ := newControlledMigrator(t, 1)
	driver.closeErr = errors.New("close failure")
	require.ErrorIs(t, m.Close(), driver.closeErr)
	require.ErrorIs(t, m.Close(), driver.closeErr)
	require.Equal(t, int32(1), driver.closes.Load())

	files := &closeableFS{FS: stepMigrations()}
	borrower, _ := newSQLiteMigrator(t, files, ".")
	require.NoError(t, borrower.Up(t.Context()))
	require.NoError(t, borrower.Close())
	require.False(t, files.closed)
}

type closeableFS struct {
	fs.FS
	closed bool
}

func (f *closeableFS) Close() error { f.closed = true; return nil }

var controlledDriverID atomic.Uint64

// controlledDriver is an external database boundary with a joinable first SQL call.
type controlledDriver struct {
	*stub.Stub
	started  chan struct{}
	release  chan struct{}
	firstRun sync.Once
	closes   atomic.Int32
	runErr   error
	closeErr error
}

func (d *controlledDriver) Open(_ string) (database.Driver, error) { return d, nil }

func (d *controlledDriver) Run(reader io.Reader) error {
	d.firstRun.Do(func() { close(d.started); <-d.release })
	if d.runErr != nil {
		return d.runErr
	}
	return d.Stub.Run(reader)
}

func (d *controlledDriver) Close() error {
	d.closes.Add(1)
	return d.closeErr
}

func newControlledMigrator(t *testing.T, count int) (*dbmigrate.Migrator, *controlledDriver, func()) {
	t.Helper()
	driver := &controlledDriver{
		Stub:    &stub.Stub{CurrentVersion: database.NilVersion},
		started: make(chan struct{}), release: make(chan struct{}),
	}
	scheme := fmt.Sprintf("controlled%d", controlledDriverID.Add(1))
	database.Register(scheme, driver)
	files := make(fstest.MapFS)
	for n := 1; n <= count; n++ {
		for _, direction := range []string{"up", "down"} {
			files[fmt.Sprintf("%d_step.%s.sql", n, direction)] = &fstest.MapFile{Data: []byte("SELECT 1;")}
		}
	}
	m, err := dbmigrate.New(files, dbmigrate.Config{DatabaseURL: scheme + "://", Directories: []string{"."}})
	require.NoError(t, err)
	t.Cleanup(func() { _ = m.Close() })
	release := sync.OnceFunc(func() { close(driver.release) })
	t.Cleanup(release)
	return m, driver, release
}

func waitStarted(t *testing.T, started <-chan struct{}) {
	t.Helper()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("migration did not start")
	}
}

func assertStillRunning(t *testing.T, result <-chan error) {
	t.Helper()
	select {
	case err := <-result:
		t.Fatalf("operation returned before SQL completed: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
}

func waitResult(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("operation did not finish")
		return nil
	}
}
