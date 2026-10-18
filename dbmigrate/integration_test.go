//go:build integration

package dbmigrate_test

import (
	"database/sql"
	"fmt"
	"net"
	"testing"
	"testing/fstest"
	"time"

	_ "github.com/ClickHouse/clickhouse-go" // Register the SQL driver required by the upstream ClickHouse adapter.
	"github.com/devctllabs/go-libs/dbmigrate"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse" // Register the migration driver.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"   // Register the migration and SQL drivers.
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresMigrations(t *testing.T) {
	t.Parallel()
	container, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:17.5-alpine",
			Env:          map[string]string{"POSTGRES_DB": "app", "POSTGRES_USER": "app", "POSTGRES_PASSWORD": "password"},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Minute),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(container)) })
	address := containerAddress(t, container, "5432/tcp")
	dsn := fmt.Sprintf("postgres://app:password@%s/app?sslmode=disable", address)
	testMigrationBackend(t, migrationBackend{
		migrationURL: dsn, sqlDriver: "postgres", sqlDSN: dsn,
		tableDDL:   "CREATE TABLE %s (id INTEGER PRIMARY KEY)",
		tableQuery: "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1",
	})
}

func TestClickHouseMigrations(t *testing.T) {
	t.Parallel()
	container, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "clickhouse/clickhouse-server:25.3-alpine",
			Env: map[string]string{
				"CLICKHOUSE_DB": "app", "CLICKHOUSE_USER": "app", "CLICKHOUSE_PASSWORD": "password",
				"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT": "1",
			},
			ExposedPorts: []string{"9000/tcp", "8123/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("9000/tcp"),
				wait.ForHTTP("/ping").WithPort("8123/tcp"),
			).WithDeadline(2 * time.Minute),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(container)) })
	address := containerAddress(t, container, "9000/tcp")
	query := "?username=app&password=password&database=app"
	testMigrationBackend(t, migrationBackend{
		migrationURL: "clickhouse://" + address + query,
		sqlDriver:    "clickhouse", sqlDSN: "tcp://" + address + query,
		tableDDL:   "CREATE TABLE %s (id Int64) ENGINE = MergeTree ORDER BY id",
		tableQuery: "SELECT COUNT(*) FROM system.tables WHERE database = currentDatabase() AND name = ?",
	})
}

func containerAddress(t *testing.T, container testcontainers.Container, port string) string {
	t.Helper()
	host, err := container.Host(t.Context())
	require.NoError(t, err)
	mapped, err := container.MappedPort(t.Context(), port)
	require.NoError(t, err)
	return net.JoinHostPort(host, mapped.Port())
}

type migrationBackend struct {
	migrationURL string
	sqlDriver    string
	sqlDSN       string
	tableDDL     string
	tableQuery   string
}

func testMigrationBackend(t *testing.T, backend migrationBackend) {
	t.Helper()
	m, err := dbmigrate.New(fstest.MapFS{
		"common/nested/1_entries.up.sql":   {Data: []byte(fmt.Sprintf(backend.tableDDL, "entries"))},
		"common/nested/1_entries.down.sql": {Data: []byte("DROP TABLE entries")},
		"custom/2_audit.up.sql":            {Data: []byte(fmt.Sprintf(backend.tableDDL, "audit"))},
		"custom/2_audit.down.sql":          {Data: []byte("DROP TABLE audit")},
		"unselected/3_invalid.up.sql":      {Data: []byte("INVALID SQL")},
	}, dbmigrate.Config{DatabaseURL: backend.migrationURL, Directories: []string{"custom", "common"}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, m.Close()) })
	db, err := sql.Open(backend.sqlDriver, backend.sqlDSN)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	assertTable := func(name string, expected int) {
		t.Helper()
		var count int
		require.NoError(t, db.QueryRowContext(t.Context(), backend.tableQuery, name).Scan(&count))
		require.Equal(t, expected, count)
	}
	require.NoError(t, m.UpN(t.Context(), 1))
	assertVersion(t, m, 1, false)
	assertTable("entries", 1)
	assertTable("audit", 0)
	require.NoError(t, m.Up(t.Context()))
	assertVersion(t, m, 2, false)
	assertTable("audit", 1)
	require.ErrorIs(t, m.Up(t.Context()), migrate.ErrNoChange)
	require.NoError(t, m.Down(t.Context(), 1))
	assertVersion(t, m, 1, false)
	assertTable("audit", 0)
	require.NoError(t, m.Up(t.Context()))
	require.NoError(t, m.Down(t.Context(), 2))
	_, _, err = m.Version()
	require.ErrorIs(t, err, migrate.ErrNilVersion)
	assertTable("entries", 0)
	assertTable("audit", 0)
}
