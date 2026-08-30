//go:build integration

package postgresdb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/devctllabs/go-libs/postgresdb"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestPostgresEndpointAndTransactionBehavior(t *testing.T) {
	ctx := context.Background()
	container, err := tcpostgres.Run(
		ctx,
		"postgres:17.5-alpine",
		tcpostgres.WithDatabase("app"),
		tcpostgres.WithUsername("app"),
		tcpostgres.WithPassword("password"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(container)) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := postgresdb.Open(ctx, postgresdb.Config{
		Writer: postgresdb.EndpointConfig{DSN: dsn},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.Writer().Check(ctx))
	require.NoError(t, db.Reader().Check(ctx))
	require.False(t, db.Writer().InTransaction(ctx))
	_, err = db.Writer().Exec(ctx, `CREATE TABLE entries (value text NOT NULL)`)
	require.NoError(t, err)

	callbackErr := errors.New("rollback requested")
	err = db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		require.True(t, db.Writer().InTransaction(txCtx))
		require.True(t, db.Reader().InTransaction(txCtx))
		_, insertErr := db.Writer().Exec(txCtx, `INSERT INTO entries VALUES ($1)`, "pending")
		require.NoError(t, insertErr)
		var count int
		queryErr := db.Reader().QueryRow(txCtx, `SELECT COUNT(*) FROM entries`).Scan(&count)
		require.NoError(t, queryErr)
		require.Equal(t, 1, count)
		return callbackErr
	})
	require.ErrorIs(t, err, callbackErr)

	var count int
	err = db.Reader().QueryRow(ctx, `SELECT COUNT(*) FROM entries`).Scan(&count)
	require.NoError(t, err)
	require.Zero(t, count)
	require.Same(t, db.Reader(), db.TxManagers().Reader())
	require.Same(t, db.Writer(), db.TxManagers().Writer())
}

func TestPgBouncerTransactionPooling(t *testing.T) {
	ctx := context.Background()
	dsn := startPgBouncer(t, ctx)
	db, err := postgresdb.Open(ctx, postgresdb.Config{
		Writer: postgresdb.EndpointConfig{DSN: dsn},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.Writer().Check(ctx))
	_, err = db.Writer().Exec(ctx, `CREATE TABLE values_table (value integer NOT NULL)`)
	require.NoError(t, err)

	for value := 1; value <= 3; value++ {
		var selected int
		err = db.Reader().QueryRow(ctx, `SELECT $1::integer`, value).Scan(&selected)
		require.NoError(t, err)
		require.Equal(t, value, selected)
	}
	err = db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		_, insertErr := db.Writer().Exec(txCtx, `INSERT INTO values_table VALUES ($1)`, 42)
		return insertErr
	})
	require.NoError(t, err)
	var count int
	err = db.Reader().QueryRow(ctx, `SELECT COUNT(*) FROM values_table`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestPostgresTelemetryUsesExplicitProviderAndProtectsQueryData(t *testing.T) {
	ctx := context.Background()
	dsn := startPostgres(t, ctx)
	globalRecorder := tracetest.NewSpanRecorder()
	globalProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(globalRecorder))
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(globalProvider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		require.NoError(t, globalProvider.Shutdown(context.Background()))
	})
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	db, err := postgresdb.Open(ctx, postgresdb.Config{
		Writer: postgresdb.EndpointConfig{DSN: dsn},
		Telemetry: postgresdb.Telemetry{
			TracerProvider: provider,
		},
	})
	require.NoError(t, err)
	var value string
	queryCtx, parentSpan := provider.Tracer("postgresdb-integration-test").Start(ctx, "request")
	err = db.Writer().QueryRow(queryCtx, `SELECT $1::text`, "sensitive-argument").Scan(&value)
	require.NoError(t, err)
	parentSpan.End()
	require.NoError(t, db.Close())
	serialized := serializePostgresSpans(recorder.Ended())
	require.NotEmpty(t, recorder.Ended())
	require.Empty(t, globalRecorder.Ended())
	require.NotContains(t, serialized, "SELECT $1::text")
	require.NotContains(t, serialized, "sensitive-argument")

	visibleRecorder := tracetest.NewSpanRecorder()
	visibleProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(visibleRecorder))
	t.Cleanup(func() { require.NoError(t, visibleProvider.Shutdown(context.Background())) })
	visibleDB, err := postgresdb.Open(ctx, postgresdb.Config{
		Writer: postgresdb.EndpointConfig{DSN: dsn},
		Telemetry: postgresdb.Telemetry{
			TracerProvider:   visibleProvider,
			IncludeQueryText: true,
		},
	})
	require.NoError(t, err)
	queryCtx, parentSpan = visibleProvider.Tracer("postgresdb-integration-test").Start(ctx, "request")
	err = visibleDB.Writer().QueryRow(
		queryCtx,
		`SELECT $1::text AS visible_statement`,
		"sensitive-argument",
	).Scan(&value)
	require.NoError(t, err)
	parentSpan.End()
	require.NoError(t, visibleDB.Close())
	serialized = serializePostgresSpans(visibleRecorder.Ended())
	require.Contains(t, serialized, "SELECT $1::text AS visible_statement")
	require.NotContains(t, serialized, "sensitive-argument")
}

func startPostgres(t *testing.T, ctx context.Context) string {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx,
		"postgres:17.5-alpine",
		tcpostgres.WithDatabase("app"),
		tcpostgres.WithUsername("app"),
		tcpostgres.WithPassword("password"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(container)) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	return dsn
}

func serializePostgresSpans(spans []sdktrace.ReadOnlySpan) string {
	var serialized strings.Builder
	for _, span := range spans {
		_, _ = fmt.Fprintln(&serialized, span.Name())
		for _, attribute := range span.Attributes() {
			_, _ = fmt.Fprintf(&serialized, "%s=%v\n", attribute.Key, attribute.Value.AsInterface())
		}
	}
	return serialized.String()
}

func startPgBouncer(t *testing.T, ctx context.Context) string {
	t.Helper()
	dockerNetwork, err := network.New(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, dockerNetwork.Remove(context.Background())) })
	postgresContainer, err := tcpostgres.Run(
		ctx,
		"postgres:17.5-alpine",
		tcpostgres.WithDatabase("app"),
		tcpostgres.WithUsername("app"),
		tcpostgres.WithPassword("password"),
		tcpostgres.BasicWaitStrategies(),
		network.WithNetworkName([]string{"postgres"}, dockerNetwork.Name),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(postgresContainer)) })
	pgBouncer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:          "edoburu/pgbouncer:v1.24.1-p0",
			ExposedPorts:   []string{"5432/tcp"},
			Networks:       []string{dockerNetwork.Name},
			NetworkAliases: map[string][]string{dockerNetwork.Name: {"pgbouncer"}},
			Env: map[string]string{
				"DB_HOST":           "postgres",
				"DB_PORT":           "5432",
				"DB_USER":           "app",
				"DB_PASSWORD":       "password",
				"DB_NAME":           "app",
				"POOL_MODE":         "transaction",
				"MAX_CLIENT_CONN":   "100",
				"DEFAULT_POOL_SIZE": "10",
				"AUTH_TYPE":         "scram-sha-256",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(pgBouncer)) })
	host, err := pgBouncer.Host(ctx)
	require.NoError(t, err)
	port, err := pgBouncer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	return fmt.Sprintf("postgres://app:password@%s:%s/app?sslmode=disable", host, port.Port())
}
