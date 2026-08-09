package filesystem

import (
	"context"
	"io/fs"
)

//go:generate go tool mockgen -destination mocks/filesystem.gen.go -package mocks . Copier,Merger,Writer,Remover

// Copier copies filesystem trees without overwriting existing files. It is a
// narrow infrastructure capability; application services should normally use a
// consumer-owned interface that describes their storage operation.
type Copier interface {
	// Copy copies source into destination. Existing directories are reused, but
	// an existing file or symbolic link causes an error matching fs.ErrExist.
	// A nil source or invalid destination causes an error matching fs.ErrInvalid.
	// The operation honors ctx but is not atomic and may leave a partial result
	// on failure or cancellation.
	Copy(ctx context.Context, source fs.FS, destination string) error
}

// Merger merges filesystem trees while preserving compatible destination
// entries.
type Merger interface {
	// Merge copies source into destination. Regular files and symbolic links of
	// the same type are replaced, directories are reused, and unrelated entries
	// are preserved. A type mismatch causes an error matching fs.ErrExist. The
	// operation honors ctx but is not atomic and may leave a partial result on
	// failure or cancellation. A nil source or invalid destination causes an
	// error matching fs.ErrInvalid.
	Merge(ctx context.Context, source fs.FS, destination string) error
}

// Writer creates directories and writes files within a filesystem root.
type Writer interface {
	// MkdirAll creates name and any missing parents with perm, subject to the
	// process umask. Name must satisfy fs.ValidPath. A canceled ctx is returned
	// before the filesystem is changed.
	MkdirAll(ctx context.Context, name string, perm fs.FileMode) error
	// WriteFile writes data to name, creating or truncating the file. It does
	// not create missing parent directories. Name must satisfy fs.ValidPath. A
	// canceled ctx is returned before the filesystem is changed.
	WriteFile(ctx context.Context, name string, data []byte, perm fs.FileMode) error
}

// Remover removes entries within a filesystem root.
type Remover interface {
	// Remove removes the file or empty directory named name. Name must satisfy
	// fs.ValidPath. A canceled ctx is returned before the filesystem is changed.
	Remove(ctx context.Context, name string) error
}
