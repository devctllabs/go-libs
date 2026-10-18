# dbmigrate

`dbmigrate` runs recursively selected SQL migrations through
[`golang-migrate`](https://github.com/golang-migrate/migrate). PostgreSQL,
ClickHouse, and SQLite use the same API and upstream database adapters.

## Usage

```go
import (
    "context"
    "errors"
    "io/fs"

    "github.com/devctllabs/go-libs/dbmigrate"
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
)

func apply(ctx context.Context, files fs.FS, databaseURL string) (err error) {
    m, err := dbmigrate.New(files, dbmigrate.Config{
        DatabaseURL: databaseURL,
        Directories: []string{"common", "dev"},
    })
    if err != nil {
        return err
    }
    defer func() { err = errors.Join(err, m.Close()) }()

    err = m.Up(ctx)
    if errors.Is(err, migrate.ErrNoChange) {
        return nil // This application treats an up-to-date database as success.
    }
    return err
}
```

Pass `os.DirFS("migrations/primary/postgres")` for disk files. For embedded
migrations, pass an `embed.FS`, using `fs.Sub` when necessary to select its root.
See [example_test.go](example_test.go) for an executable embedded SQLite example.

The application selects its adapters with blank imports:

| Database | Migration adapter | URL scheme |
| --- | --- | --- |
| PostgreSQL | `github.com/golang-migrate/migrate/v4/database/postgres` | `postgres://` or `postgresql://` |
| ClickHouse | `github.com/golang-migrate/migrate/v4/database/clickhouse` | `clickhouse://` |
| SQLite, pure Go | `github.com/golang-migrate/migrate/v4/database/sqlite` | `sqlite://` |

ClickHouse additionally requires the SQL driver import
`_ "github.com/ClickHouse/clickhouse-go"`. It is the v1 driver used by the upstream
migration adapter; the application's runtime ClickHouse client can be separate.
Driver settings, including the migration table and multi-statement behavior, use
the upstream URL query parameters. There is no hardcoded database allowlist.

## Directory selection and versions

For example:

```text
migrations/primary/postgres/
├── common/
│   └── schema/
│       ├── 20261018090000_create_entries.up.sql
│       └── 20261018090000_create_entries.down.sql
├── dev/
│   ├── 20261018090100_seed_entries.up.sql
│   └── 20261018090100_seed_entries.down.sql
├── uat/
└── prod/
```

`Directories: []string{"common", "dev"}` selects those trees recursively.
Directory names are arbitrary; no folder is automatically included. Paths are
relative `io/fs` paths, and the list must be nonempty. `[]string{"."}` explicitly
selects the whole source. Each selected path must exist and be a directory.

All selected files form **one history** in the driver's migration table, ordered
by numeric version. Directory order and nesting do not affect execution. Repeated
or overlapping selections include the same path once. Different files with the
same version and direction return upstream `source.ErrDuplicateMigration` with
the conflicting paths.

Only `.sql` files are considered. Filenames use upstream parsing:
`<version>_<name>.up.sql` or `<version>_<name>.down.sql`. Timestamp names match
Devctl's convention; incrementing numeric versions are also accepted. Invalid
names are ignored, as with the upstream filesystem source. SQL is passed through
without parsing, template expansion, or rewriting.

Up/down counterparts are optional, as in golang-migrate. A missing direction
executes an empty step and still changes the recorded version; a successful down
step therefore does not prove that SQL changes were reversed.

The file index is fixed by `New`; file contents are opened when requested by the
engine. Keep the filesystem available and migrations unchanged during operations.
For each database, keep directory selection stable, preserve applied files, and
assign new migrations versions above its current version. For example, adding
`common/150` after applying `dev/200` will not apply version 150. The library does
not keep checksums or a separate applied-file journal.

## Operations and lifetime

- `Up(ctx)` applies all pending migrations.
- `UpN(ctx, n)` and `Down(ctx, n)` require `n > 0`. Their upstream `Steps` behavior
  can apply available steps before returning an error when fewer than `n` exist.
- `Version()` returns version and dirty state, or `migrate.ErrNilVersion` before
  any migration has been applied or after a complete rollback.
- `Force(version)` accepts versions `>= -1`, clears dirty, and executes no SQL.
  Repair the actual database before using it for recovery.
- `Close()` waits for the active call and closes the database once. It never
  closes the caller's filesystem; repeated calls preserve the first close result.

**Context is checked only before execution.** A canceled context prevents an
operation from starting. Once started, the entire upstream operation runs to
completion, even if the context is canceled; its original result is returned.
`GracefulStop` is intentionally unused because golang-migrate v4.19.1 has a data
race between its migration reader and executor during graceful stopping. Configure
connection and statement timeouts through the selected database driver.

Only one operation runs on an instance at a time; concurrent calls return
`migrate.ErrLocked`. After closing, operations return `ErrClosed`.
Invalid construction arguments return `ErrInvalidConfig`, and nonpositive step
counts return `ErrInvalidSteps`. Filesystem and upstream errors retain their causes
for `errors.Is`/`errors.As`. In particular, `ErrNoChange` is not automatically
converted to success. An empty indexed source preserves upstream `fs.ErrNotExist`.

Transactions, dirty-state handling, and database locks belong to the upstream
adapter. ClickHouse migrations are non-transactional, and lock guarantees vary by
driver; the wrapper adds no cross-process locking. Use this library in explicit
migration tooling or jobs rather than implicitly applying schemas in repository
constructors.

## Verification

```sh
go test -race ./dbmigrate/...
mise run dbmigrate:test-integration
```

SQLite tests use temporary databases. Integration tests require Docker and run
against PostgreSQL and ClickHouse containers. To check the module independently:

```sh
cd dbmigrate
GOWORK=off go mod tidy -diff
GOWORK=off go test -race ./...
```
