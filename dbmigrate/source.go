package dbmigrate

import (
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
)

type recursiveSource struct {
	fsys       fs.FS
	migrations *source.Migrations
}

func newSource(fsys fs.FS, directories []string) (*recursiveSource, error) {
	s := &recursiveSource{fsys: fsys, migrations: source.NewMigrations()}
	seen := make(map[string]bool)
	for _, directory := range directories {
		info, err := fs.Stat(fsys, directory)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%w: %q is not a directory", ErrInvalidConfig, directory)
		}
		err = fs.WalkDir(fsys, directory, func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || seen[name] || !strings.HasSuffix(name, ".sql") {
				return nil
			}
			seen[name] = true
			return s.append(name, entry)
		})
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *recursiveSource) append(name string, entry fs.DirEntry) error {
	migration, err := source.DefaultParse(entry.Name())
	if err != nil {
		return nil // Match the upstream filesystem source's handling of invalid names.
	}
	migration.Raw = name
	if s.migrations.Append(migration) {
		return nil
	}
	// These upstream methods look up indexed files; they do not execute SQL.
	var existing *source.Migration
	if migration.Direction == source.Up {
		existing, _ = s.migrations.Up(migration.Version)
	} else {
		existing, _ = s.migrations.Down(migration.Version)
	}
	info, err := entry.Info()
	if err != nil {
		return err
	}
	return fmt.Errorf("migration %q conflicts with %q: %w", name, existing.Raw,
		source.ErrDuplicateMigration{Migration: *migration, FileInfo: info})
}

func (*recursiveSource) Open(_ string) (source.Driver, error) {
	return nil, fmt.Errorf("recursive migration sources require an fs.FS")
}

func (*recursiveSource) Close() error { return nil }

func (s *recursiveSource) First() (uint, error) {
	return sourceVersion(s.migrations.First())
}

func (s *recursiveSource) Prev(version uint) (uint, error) {
	return sourceVersion(s.migrations.Prev(version))
}

func (s *recursiveSource) Next(version uint) (uint, error) {
	return sourceVersion(s.migrations.Next(version))
}

func sourceVersion(version uint, exists bool) (uint, error) {
	if !exists {
		return 0, fs.ErrNotExist
	}
	return version, nil
}

func (s *recursiveSource) ReadUp(version uint) (io.ReadCloser, string, error) {
	return s.read(s.migrations.Up(version))
}

func (s *recursiveSource) ReadDown(version uint) (io.ReadCloser, string, error) {
	return s.read(s.migrations.Down(version))
}

func (s *recursiveSource) read(migration *source.Migration, exists bool) (io.ReadCloser, string, error) {
	if !exists {
		return nil, "", fs.ErrNotExist
	}
	file, err := s.fsys.Open(migration.Raw)
	if err != nil {
		return nil, "", fmt.Errorf("read migration %q: %w", migration.Raw, err)
	}
	return file, migration.Identifier, nil
}
