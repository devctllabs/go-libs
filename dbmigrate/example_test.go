package dbmigrate_test

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/devctllabs/go-libs/dbmigrate"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite" // Register the caller-selected migration adapter.
)

//go:embed testdata/migrations
var embeddedMigrations embed.FS

func ExampleMigrator() {
	directory, err := os.MkdirTemp("", "dbmigrate-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(directory)
	files, err := fs.Sub(embeddedMigrations, "testdata/migrations")
	if err != nil {
		panic(err)
	}
	// The application also imports the selected upstream database adapter.
	m, err := dbmigrate.New(files, dbmigrate.Config{
		DatabaseURL: "sqlite://" + filepath.Join(directory, "app.db"),
		Directories: []string{"common", "dev"},
	})
	if err != nil {
		panic(err)
	}
	defer m.Close()
	if err := m.Up(context.Background()); err != nil {
		panic(err)
	}
	version, dirty, err := m.Version()
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Println(version, dirty)
	// Output: 20261018090100 false
}
