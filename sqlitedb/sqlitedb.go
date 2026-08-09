package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/devctllabs/go-libs/txmanager"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	// Register the modernc SQLite driver used by Open.
	_ "modernc.org/sqlite"
)

// Config defines one SQLite database instance.
type Config struct {
	DSN         string
	BusyTimeout time.Duration
	ReaderPool  PoolConfig
	Telemetry   Telemetry
}

// Telemetry supplies instance-scoped OpenTelemetry dependencies.
type Telemetry struct {
	TracerProvider   trace.TracerProvider
	MeterProvider    metric.MeterProvider
	IncludeQueryText bool
}

// PoolConfig controls the SQLite reader connection pool.
type PoolConfig struct {
	MaxOpenConnections    int
	MaxIdleConnections    int
	ConnectionMaxLifetime time.Duration
	ConnectionMaxIdleTime time.Duration
}

// DB owns the SQLite reader and writer resources.
type DB struct {
	reader     *Endpoint
	writer     *Endpoint
	managers   txmanager.Managers
	readerPool *sql.DB
	writerPool *sql.DB
	metrics    []metric.Registration
	closeOnce  sync.Once
	closeErr   error
}

// Open constructs a SQLite database instance.
func Open(ctx context.Context, cfg Config) (*DB, error) {
	if ctx == nil {
		return nil, errors.New("sqlitedb: context must not be nil")
	}
	if strings.TrimSpace(cfg.DSN) == "" {
		return nil, errors.New("sqlitedb: DSN must not be blank")
	}
	busyTimeout, readerPoolConfig, err := validateConfig(cfg)
	if err != nil {
		return nil, err
	}
	writerDSN, readerDSN, memory, err := connectionDSNs(cfg.DSN, busyTimeout)
	if err != nil {
		return nil, err
	}
	resources := &sqliteResources{}
	if err := resources.openWriter(ctx, cfg.Telemetry, writerDSN, memory); err != nil {
		return nil, resources.cleanup(err)
	}
	if err := resources.openReader(ctx, cfg.Telemetry, readerDSN, readerPoolConfig); err != nil {
		return nil, resources.cleanup(err)
	}
	db, err := resources.buildDB()
	if err != nil {
		return nil, resources.cleanup(err)
	}
	return db, nil
}

type sqliteResources struct {
	readerPool    *sql.DB
	writerPool    *sql.DB
	readerMetrics metric.Registration
	writerMetrics metric.Registration
}

func (r *sqliteResources) openWriter(ctx context.Context, telemetry Telemetry, dsn string, memory bool) error {
	writerOptions, writerMetricOptions := telemetryOptions(telemetry, "writer")
	writerPool, err := otelsql.Open("sqlite", dsn, writerOptions...)
	if err != nil {
		return fmt.Errorf("sqlitedb: open writer: %w", err)
	}
	r.writerPool = writerPool
	writerPool.SetMaxOpenConns(1)
	writerPool.SetMaxIdleConns(1)
	if err := writerPool.PingContext(ctx); err != nil {
		return fmt.Errorf("sqlitedb: ping writer: %w", err)
	}
	if err := enableWAL(ctx, writerPool, memory); err != nil {
		return err
	}
	r.writerMetrics, err = otelsql.RegisterDBStatsMetrics(writerPool, writerMetricOptions...)
	if err != nil {
		return fmt.Errorf("sqlitedb: register writer metrics: %w", err)
	}
	return nil
}

func enableWAL(ctx context.Context, writerPool *sql.DB, memory bool) error {
	if memory {
		return nil
	}
	var journalMode string
	if err := writerPool.QueryRowContext(ctx, `PRAGMA journal_mode=WAL`).Scan(&journalMode); err != nil {
		return fmt.Errorf("sqlitedb: enable WAL: %w", err)
	}
	if !strings.EqualFold(journalMode, "wal") {
		return fmt.Errorf("sqlitedb: WAL is unavailable, journal mode is %q", journalMode)
	}
	return nil
}

func (r *sqliteResources) openReader(ctx context.Context, telemetry Telemetry, dsn string, poolConfig PoolConfig) error {
	readerOptions, readerMetricOptions := telemetryOptions(telemetry, "reader")
	readerPool, err := otelsql.Open("sqlite", dsn, readerOptions...)
	if err != nil {
		return fmt.Errorf("sqlitedb: open reader: %w", err)
	}
	r.readerPool = readerPool
	readerPool.SetMaxOpenConns(poolConfig.MaxOpenConnections)
	readerPool.SetMaxIdleConns(poolConfig.MaxIdleConnections)
	readerPool.SetConnMaxLifetime(poolConfig.ConnectionMaxLifetime)
	readerPool.SetConnMaxIdleTime(poolConfig.ConnectionMaxIdleTime)
	if err := readerPool.PingContext(ctx); err != nil {
		return fmt.Errorf("sqlitedb: ping reader: %w", err)
	}
	r.readerMetrics, err = otelsql.RegisterDBStatsMetrics(readerPool, readerMetricOptions...)
	if err != nil {
		return fmt.Errorf("sqlitedb: register reader metrics: %w", err)
	}
	return nil
}

func (r *sqliteResources) buildDB() (*DB, error) {
	backend := &transactionBackend{reader: r.readerPool, writer: r.writerPool}
	coordinator, err := txmanager.NewCoordinator[*sql.Tx](backend)
	if err != nil {
		return nil, err
	}
	reader := &Endpoint{pool: r.readerPool, coordinator: coordinator, manager: coordinator.Reader()}
	writer := &Endpoint{pool: r.writerPool, coordinator: coordinator, manager: coordinator.Writer()}
	managers, err := txmanager.NewManagers(reader, writer)
	if err != nil {
		return nil, err
	}
	return &DB{
		reader: reader, writer: writer, managers: managers,
		readerPool: r.readerPool, writerPool: r.writerPool,
		metrics: []metric.Registration{r.readerMetrics, r.writerMetrics},
	}, nil
}

func (r *sqliteResources) cleanup(cause error) error {
	errorsToJoin := []error{cause}
	if r.readerMetrics != nil {
		errorsToJoin = append(errorsToJoin, r.readerMetrics.Unregister())
	}
	if r.readerPool != nil {
		errorsToJoin = append(errorsToJoin, r.readerPool.Close())
	}
	if r.writerMetrics != nil {
		errorsToJoin = append(errorsToJoin, r.writerMetrics.Unregister())
	}
	if r.writerPool != nil {
		errorsToJoin = append(errorsToJoin, r.writerPool.Close())
	}
	if len(errorsToJoin) == 1 {
		return cause
	}
	return errors.Join(errorsToJoin...)
}

// Reader returns the read-only endpoint.
func (db *DB) Reader() *Endpoint {
	return db.reader
}

// Writer returns the read-write endpoint.
func (db *DB) Writer() *Endpoint {
	return db.writer
}

// TxManagers returns reader and writer transaction managers backed by this DB.
func (db *DB) TxManagers() txmanager.Managers {
	return db.managers
}

// Close releases both endpoint pools once.
func (db *DB) Close() error {
	db.closeOnce.Do(func() {
		metricErrors := make([]error, 0, len(db.metrics))
		for _, registration := range db.metrics {
			metricErrors = append(metricErrors, registration.Unregister())
		}
		db.closeErr = errors.Join(
			errors.Join(metricErrors...),
			db.readerPool.Close(),
			db.writerPool.Close(),
		)
	})
	return db.closeErr
}

func validateConfig(cfg Config) (time.Duration, PoolConfig, error) {
	busyTimeout := cfg.BusyTimeout
	if busyTimeout == 0 {
		busyTimeout = 5 * time.Second
	}
	if busyTimeout < time.Millisecond {
		return 0, PoolConfig{}, errors.New("sqlitedb: busy timeout must be at least one millisecond")
	}
	pool := cfg.ReaderPool
	if pool.MaxOpenConnections < 0 || pool.MaxIdleConnections < 0 ||
		pool.ConnectionMaxLifetime < 0 || pool.ConnectionMaxIdleTime < 0 {
		return 0, PoolConfig{}, errors.New("sqlitedb: reader pool values must not be negative")
	}
	if pool.MaxOpenConnections == 0 {
		pool.MaxOpenConnections = 1
	}
	if pool.MaxIdleConnections == 0 {
		pool.MaxIdleConnections = 1
	}
	if pool.MaxIdleConnections > pool.MaxOpenConnections {
		return 0, PoolConfig{}, errors.New("sqlitedb: reader max idle connections exceeds max open connections")
	}
	return busyTimeout, pool, nil
}

func connectionDSNs(rawDSN string, busyTimeout time.Duration) (string, string, bool, error) {
	if rawDSN == ":memory:" {
		return "", "", false, errors.New("sqlitedb: bare :memory: DSN is not supported")
	}
	parsed, err := sqliteURI(rawDSN)
	if err != nil {
		return "", "", false, err
	}
	writerValues := parsed.Query()
	memory := writerValues.Get("mode") == "memory"
	if memory {
		if writerValues.Get("cache") != "shared" || parsed.Opaque == "" && parsed.Path == "" {
			return "", "", false, errors.New("sqlitedb: memory DSN must be named and use cache=shared")
		}
	} else {
		writerValues.Set("mode", "rwc")
	}
	writerValues.Set("_txlock", "immediate")
	writerValues.Add("_pragma", "foreign_keys(1)")
	writerValues.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busyTimeout.Milliseconds()))
	parsed.RawQuery = writerValues.Encode()
	writerDSN := parsed.String()

	readerValues := cloneValues(writerValues)
	if !memory {
		readerValues.Set("mode", "ro")
	}
	readerValues.Del("_txlock")
	readerValues.Add("_pragma", "query_only(1)")
	parsed.RawQuery = readerValues.Encode()
	return writerDSN, parsed.String(), memory, nil
}

func sqliteURI(rawDSN string) (*url.URL, error) {
	if strings.HasPrefix(rawDSN, "file:") {
		parsed, err := url.Parse(rawDSN)
		if err != nil {
			return nil, fmt.Errorf("sqlitedb: parse DSN: %w", err)
		}
		return parsed, nil
	}
	return &url.URL{Scheme: "file", Path: rawDSN}, nil
}

func cloneValues(values url.Values) url.Values {
	cloned := make(url.Values, len(values))
	for key, entries := range values {
		cloned[key] = append([]string(nil), entries...)
	}
	return cloned
}
