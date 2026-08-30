package kafkaoutbox

import (
	"embed"
	"io/fs"
)

//go:embed migrations/polling/*.sql migrations/cdc/*.sql
var migrationFiles embed.FS

// PollingMigrations returns versioned SQL migrations for the polling profile.
func PollingMigrations() fs.FS {
	migrations, err := fs.Sub(migrationFiles, "migrations/polling")
	if err != nil {
		panic(err)
	}
	return migrations
}

// CDCMigrations returns versioned SQL migrations for the Debezium CDC profile.
func CDCMigrations() fs.FS {
	migrations, err := fs.Sub(migrationFiles, "migrations/cdc")
	if err != nil {
		panic(err)
	}
	return migrations
}
