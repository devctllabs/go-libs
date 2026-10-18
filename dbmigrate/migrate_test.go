package dbmigrate_test

import (
	"database/sql"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/devctllabs/go-libs/dbmigrate"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite" // Register the caller-selected database driver.
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/stretchr/testify/require"
)

func TestMigratorUp(t *testing.T) {
	t.Parallel()
	databasePath := filepath.Join(t.TempDir(), "app.db")
	m, err := dbmigrate.New(fstest.MapFS{
		"common/1_create.up.sql": {Data: []byte("CREATE TABLE entries (value TEXT);")},
	}, dbmigrate.Config{
		DatabaseURL: "sqlite://" + databasePath,
		Directories: []string{"common"},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, m.Close()) })
	require.NoError(t, m.Up(t.Context()))
	version, dirty, err := m.Version()
	require.NoError(t, err)
	require.Equal(t, uint(1), version)
	require.False(t, dirty)

	db, err := sql.Open("sqlite", databasePath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	var name string
	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT name FROM sqlite_master WHERE name = 'entries'").Scan(&name))
	require.Equal(t, "entries", name)
}

func TestMigratorSelection(t *testing.T) {
	t.Parallel()
	for _, directories := range [][]string{
		{"shared", "local"},
		{"local", "shared"},
		{"shared/base", "local", "shared", "shared"},
	} {
		t.Run(directories[0], func(t *testing.T) {
			t.Parallel()
			m, db := newSQLiteMigrator(t, fstest.MapFS{
				"shared/base/1_create.up.sql": {Data: []byte("CREATE TABLE entries (value TEXT);")},
				"local/deep/2_insert.up.sql":  {Data: []byte("INSERT INTO entries VALUES ('two');")},
				"shared/10_insert.up.sql":     {Data: []byte("INSERT INTO entries VALUES ('ten');")},
				"production/3_invalid.up.sql": {Data: []byte("INVALID SQL")},
				"shared/ignored.sql":          {Data: []byte("INVALID SQL")},
				"shared/11_ignore.up.txt":     {Data: []byte("INVALID SQL")},
			}, directories...)
			require.NoError(t, m.Up(t.Context()))
			rows, err := db.QueryContext(t.Context(), "SELECT value FROM entries ORDER BY rowid")
			require.NoError(t, err)
			defer rows.Close()
			var values []string
			for rows.Next() {
				var value string
				require.NoError(t, rows.Scan(&value))
				values = append(values, value)
			}
			require.NoError(t, rows.Err())
			require.Equal(t, []string{"two", "ten"}, values)
			version, dirty, err := m.Version()
			require.NoError(t, err)
			require.Equal(t, uint(10), version)
			require.False(t, dirty)
		})
	}
}

func TestMigratorDuplicateMigration(t *testing.T) {
	t.Parallel()
	m, err := dbmigrate.New(fstest.MapFS{
		"a/1_first.up.sql": {Data: []byte("SELECT 1;")},
		"b/1_other.up.sql": {Data: []byte("SELECT 2;")},
	}, dbmigrate.Config{DatabaseURL: "unregistered://", Directories: []string{"a", "b"}})
	require.Nil(t, m)
	var duplicate source.ErrDuplicateMigration
	require.ErrorAs(t, err, &duplicate)
	require.ErrorContains(t, err, "a/1_first.up.sql")
	require.ErrorContains(t, err, "b/1_other.up.sql")
}

func newSQLiteMigrator(t *testing.T, fsys fs.FS, directories ...string) (*dbmigrate.Migrator, *sql.DB) {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), "app.db")
	m, err := dbmigrate.New(fsys, dbmigrate.Config{
		DatabaseURL: "sqlite://" + databasePath,
		Directories: directories,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, m.Close()) })
	db, err := sql.Open("sqlite", databasePath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return m, db
}

func stepMigrations() fstest.MapFS {
	return fstest.MapFS{
		"1_create.up.sql":          {Data: []byte("CREATE TABLE entries (value TEXT);")},
		"1_create.down.sql":        {Data: []byte("DROP TABLE entries;")},
		"nested/2_insert.up.sql":   {Data: []byte("INSERT INTO entries VALUES ('two');")},
		"nested/2_insert.down.sql": {Data: []byte("DELETE FROM entries WHERE value = 'two';")},
		"10_insert.up.sql":         {Data: []byte("INSERT INTO entries VALUES ('ten');")},
		"10_insert.down.sql":       {Data: []byte("DELETE FROM entries WHERE value = 'ten';")},
	}
}

func assertVersion(t *testing.T, m *dbmigrate.Migrator, expected uint, expectedDirty bool) {
	t.Helper()
	version, dirty, err := m.Version()
	require.NoError(t, err)
	require.Equal(t, expected, version)
	require.Equal(t, expectedDirty, dirty)
}

func TestMigratorSteps(t *testing.T) {
	t.Parallel()
	m, db := newSQLiteMigrator(t, stepMigrations(), ".")
	require.NoError(t, m.UpN(t.Context(), 2))
	assertVersion(t, m, 2, false)
	require.NoError(t, m.Down(t.Context(), 1))
	assertVersion(t, m, 1, false)
	var count int
	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM entries").Scan(&count))
	require.Zero(t, count)
	require.NoError(t, m.Up(t.Context()))
	assertVersion(t, m, 10, false)
	require.ErrorIs(t, m.Up(t.Context()), migrate.ErrNoChange)
	require.NoError(t, m.Down(t.Context(), 3))
	_, _, err := m.Version()
	require.ErrorIs(t, err, migrate.ErrNilVersion)
	require.NoError(t, m.Up(t.Context()))
	assertVersion(t, m, 10, false)
}

func TestMigratorStepLimits(t *testing.T) {
	t.Parallel()
	for _, direction := range []string{"up", "down"} {
		t.Run(direction, func(t *testing.T) {
			t.Parallel()
			m, _ := newSQLiteMigrator(t, stepMigrations(), ".")
			operation := m.UpN
			if direction == "down" {
				require.NoError(t, m.Up(t.Context()))
				operation = m.Down
			}
			for _, n := range []int{0, -1} {
				require.ErrorIs(t, operation(t.Context(), n), dbmigrate.ErrInvalidSteps)
			}
			var short migrate.ErrShortLimit
			require.ErrorAs(t, operation(t.Context(), 8), &short)
			require.Equal(t, uint(5), short.Short)
			if direction == "up" {
				assertVersion(t, m, 10, false)
			} else {
				_, _, err := m.Version()
				require.ErrorIs(t, err, migrate.ErrNilVersion)
			}
		})
	}
}

func TestMigratorMissingDirections(t *testing.T) {
	t.Parallel()
	m, db := newSQLiteMigrator(t, fstest.MapFS{
		"1_create.up.sql": {Data: []byte("CREATE TABLE entries (value TEXT);")},
		"2_drop.down.sql": {Data: []byte("DROP TABLE entries;")},
	}, ".")
	require.NoError(t, m.Up(t.Context()))
	assertVersion(t, m, 2, false)
	var count int
	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM entries").Scan(&count))
	require.NoError(t, m.Down(t.Context(), 1))
	assertVersion(t, m, 1, false)
	require.NoError(t, m.Down(t.Context(), 1))
	_, _, err := m.Version()
	require.ErrorIs(t, err, migrate.ErrNilVersion)
}

func TestMigratorDirtyAndForce(t *testing.T) {
	t.Parallel()
	m, db := newSQLiteMigrator(t, fstest.MapFS{
		"1_create.up.sql": {Data: []byte("CREATE TABLE entries (value TEXT);")},
		"2_fail.up.sql":   {Data: []byte("INVALID SQL")},
	}, ".")
	require.Error(t, m.Up(t.Context()))
	assertVersion(t, m, 2, true)
	var dirty migrate.ErrDirty
	require.ErrorAs(t, m.Up(t.Context()), &dirty)
	require.Equal(t, 2, dirty.Version)
	require.NoError(t, m.Force(1))
	assertVersion(t, m, 1, false)
	require.ErrorIs(t, m.Force(-2), migrate.ErrInvalidVersion)
	require.NoError(t, m.Force(-1))
	_, _, err := m.Version()
	require.ErrorIs(t, err, migrate.ErrNilVersion)
	var count int
	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM entries").Scan(&count))
}

func TestMigratorInvalidConfig(t *testing.T) {
	t.Parallel()
	files := stepMigrations()
	for _, tc := range []struct {
		name string
		fsys fs.FS
		cfg  dbmigrate.Config
	}{
		{name: "nil filesystem", cfg: dbmigrate.Config{DatabaseURL: "unregistered://", Directories: []string{"."}}},
		{name: "no directories", fsys: files, cfg: dbmigrate.Config{DatabaseURL: "unregistered://"}},
		{name: "no URL", fsys: files, cfg: dbmigrate.Config{Directories: []string{"."}}},
		{name: "parent path", fsys: files, cfg: dbmigrate.Config{DatabaseURL: "unregistered://", Directories: []string{"../outside"}}},
		{name: "file path", fsys: files, cfg: dbmigrate.Config{DatabaseURL: "unregistered://", Directories: []string{"1_create.up.sql"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, err := dbmigrate.New(tc.fsys, tc.cfg)
			require.Nil(t, m)
			require.ErrorIs(t, err, dbmigrate.ErrInvalidConfig)
		})
	}
}

func TestMigratorEmptyAndMissingSource(t *testing.T) {
	t.Parallel()
	m, _ := newSQLiteMigrator(t, fstest.MapFS{"empty/.gitkeep": {}}, "empty")
	require.ErrorIs(t, m.Up(t.Context()), fs.ErrNotExist)
	_, _, err := m.Version()
	require.ErrorIs(t, err, migrate.ErrNilVersion)
	missing, err := dbmigrate.New(fstest.MapFS{}, dbmigrate.Config{
		DatabaseURL: "unregistered://", Directories: []string{"missing"},
	})
	require.Nil(t, missing)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMigratorSourceErrors(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{"1_create.up.sql": {Data: []byte("SELECT 1;")}}
	denied := &failingFS{FS: files}
	m, err := dbmigrate.New(denied, dbmigrate.Config{DatabaseURL: "unregistered://", Directories: []string{"."}})
	require.Nil(t, m)
	require.ErrorIs(t, err, fs.ErrPermission)

	m, _ = newSQLiteMigrator(t, &failingFS{FS: files, suffix: ".sql"}, ".")
	err = m.Up(t.Context())
	require.ErrorIs(t, err, fs.ErrPermission)
	require.ErrorContains(t, err, "1_create.up.sql")
	_, _, err = m.Version()
	require.ErrorIs(t, err, migrate.ErrNilVersion)
}

type failingFS struct {
	fs.FS
	suffix string
}

func (f *failingFS) Open(name string) (fs.File, error) {
	if strings.HasSuffix(name, f.suffix) {
		return nil, fs.ErrPermission
	}
	return f.FS.Open(name)
}

func TestMigratorIndexIsFixedAtConstruction(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{"1_create.up.sql": {Data: []byte("CREATE TABLE entries (value TEXT);")}}
	m, _ := newSQLiteMigrator(t, files, ".")
	files["2_insert.up.sql"] = &fstest.MapFile{Data: []byte("INSERT INTO entries VALUES ('later');")}
	require.NoError(t, m.Up(t.Context()))
	assertVersion(t, m, 1, false)
	require.ErrorIs(t, m.Up(t.Context()), migrate.ErrNoChange)
}
