package postgresdb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/devctllabs/go-libs/txmanager"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config defines one PostgreSQL database instance.
type Config struct {
	Writer    EndpointConfig
	Reader    *EndpointConfig
	Telemetry Telemetry
}

// EndpointConfig configures one PostgreSQL pool.
type EndpointConfig struct {
	DSN  string
	Pool PoolConfig
}

// PoolConfig overrides a stable subset of pgx pool settings when fields are positive.
type PoolConfig struct {
	MaxConnections              int32
	MinIdleConnections          int32
	MaxConnectionLifetime       time.Duration
	MaxConnectionLifetimeJitter time.Duration
	MaxConnectionIdleTime       time.Duration
	HealthCheckPeriod           time.Duration
}

// DB owns PostgreSQL reader and writer resources.
type DB struct {
	reader     *Endpoint
	writer     *Endpoint
	managers   txmanager.Managers
	readerPool *pgxpool.Pool
	writerPool *pgxpool.Pool
	closeOnce  sync.Once
}

// Open constructs a PostgreSQL database instance without requiring network availability.
func Open(ctx context.Context, cfg Config) (*DB, error) {
	if ctx == nil {
		return nil, errors.New("postgresdb: context must not be nil")
	}
	telemetry := newTelemetryConfig(cfg.Telemetry)
	writerPool, err := openPool(ctx, cfg.Writer, telemetry, "writer")
	if err != nil {
		return nil, err
	}
	readerPool := writerPool
	if cfg.Reader != nil {
		readerPool, err = openPool(ctx, *cfg.Reader, telemetry, "reader")
		if err != nil {
			writerPool.Close()
			return nil, err
		}
	}
	return buildDB(readerPool, writerPool)
}

func openPool(ctx context.Context, endpoint EndpointConfig, telemetry telemetryConfig, role string) (*pgxpool.Pool, error) {
	if strings.TrimSpace(endpoint.DSN) == "" {
		return nil, fmt.Errorf("postgresdb: %s DSN must not be blank", role)
	}
	config, err := parsePoolConfig(endpoint, telemetry, role)
	if err != nil {
		return nil, fmt.Errorf("postgresdb: %s config: %w", role, err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("postgresdb: open %s pool: %w", role, err)
	}
	if err := telemetry.recordPoolStats(pool, role); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgresdb: instrument %s pool: %w", role, err)
	}
	return pool, nil
}

func buildDB(readerPool, writerPool *pgxpool.Pool) (*DB, error) {
	backend := &transactionBackend{reader: readerPool, writer: writerPool}
	coordinator, err := txmanager.NewCoordinator[pgx.Tx](backend)
	if err != nil {
		closePools(readerPool, writerPool)
		return nil, err
	}
	reader := &Endpoint{pool: readerPool, coordinator: coordinator, manager: coordinator.Reader()}
	writer := &Endpoint{pool: writerPool, coordinator: coordinator, manager: coordinator.Writer()}
	managers, err := txmanager.NewManagers(reader, writer)
	if err != nil {
		closePools(readerPool, writerPool)
		return nil, err
	}
	return &DB{
		reader:     reader,
		writer:     writer,
		managers:   managers,
		readerPool: readerPool,
		writerPool: writerPool,
	}, nil
}

// Reader returns the reader endpoint.
func (db *DB) Reader() *Endpoint {
	return db.reader
}

// Writer returns the writer endpoint.
func (db *DB) Writer() *Endpoint {
	return db.writer
}

// TxManagers returns reader and writer transaction managers backed by this DB.
func (db *DB) TxManagers() txmanager.Managers {
	return db.managers
}

// Close releases all distinct pools once.
func (db *DB) Close() error {
	db.closeOnce.Do(func() {
		if db.readerPool != db.writerPool {
			db.readerPool.Close()
		}
		db.writerPool.Close()
	})
	return nil
}

func closePools(reader *pgxpool.Pool, writer *pgxpool.Pool) {
	if reader != writer {
		reader.Close()
	}
	writer.Close()
}

func parsePoolConfig(endpoint EndpointConfig, telemetry telemetryConfig, role string) (*pgxpool.Config, error) {
	if err := validatePoolConfig(endpoint.Pool); err != nil {
		return nil, err
	}
	config, err := pgxpool.ParseConfig(endpoint.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	config.ConnConfig.Tracer = telemetry.tracer(role)
	if endpoint.Pool.MaxConnections > 0 {
		config.MaxConns = endpoint.Pool.MaxConnections
	}
	if endpoint.Pool.MinIdleConnections > 0 {
		config.MinIdleConns = endpoint.Pool.MinIdleConnections
	}
	if endpoint.Pool.MaxConnectionLifetime > 0 {
		config.MaxConnLifetime = endpoint.Pool.MaxConnectionLifetime
	}
	if endpoint.Pool.MaxConnectionLifetimeJitter > 0 {
		config.MaxConnLifetimeJitter = endpoint.Pool.MaxConnectionLifetimeJitter
	}
	if endpoint.Pool.MaxConnectionIdleTime > 0 {
		config.MaxConnIdleTime = endpoint.Pool.MaxConnectionIdleTime
	}
	if endpoint.Pool.HealthCheckPeriod > 0 {
		config.HealthCheckPeriod = endpoint.Pool.HealthCheckPeriod
	}
	if config.MinIdleConns > config.MaxConns {
		return nil, errors.New("minimum idle connections exceeds maximum connections")
	}
	return config, nil
}

func validatePoolConfig(config PoolConfig) error {
	if config.MaxConnections < 0 || config.MinIdleConnections < 0 ||
		config.MaxConnectionLifetime < 0 || config.MaxConnectionLifetimeJitter < 0 ||
		config.MaxConnectionIdleTime < 0 || config.HealthCheckPeriod < 0 {
		return errors.New("pool values must not be negative")
	}
	return nil
}
