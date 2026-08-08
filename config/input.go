package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// Input opens configuration data for a loader. Each call must return an
// independent stream when concurrent or repeated loading is supported.
type Input interface {
	// Open returns a new readable stream. The calling loader closes the stream.
	// An absent resource should return an error that matches fs.ErrNotExist so it
	// can be composed with Optional.
	Open(ctx context.Context) (io.ReadCloser, error)
}

type pathInput struct {
	path string
}

type fsInput struct {
	fsys fs.FS
	name string
}

// Path creates an Input backed by path in the operating system filesystem. The
// path is opened when Input.Open is called.
func Path(path string) Input {
	return pathInput{path: path}
}

// FromFS creates an Input backed by name in fsys. Name must satisfy
// fs.ValidPath. The file is opened when Input.Open is called.
func FromFS(fsys fs.FS, name string) Input {
	return fsInput{fsys: fsys, name: name}
}

func (i pathInput) Open(ctx context.Context) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return os.Open(i.path)
}

func (i fsInput) Open(ctx context.Context) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if i.fsys == nil {
		return nil, errors.New("config: nil filesystem")
	}

	return i.fsys.Open(i.name)
}

func readInput(ctx context.Context, input Input, kind string) ([]byte, error) {
	if isNilValue(input) {
		return nil, fmt.Errorf("config: nil %s input", kind)
	}

	reader, err := input.Open(ctx)
	if err != nil {
		return nil, fmt.Errorf("config: open %s input: %w", kind, err)
	}
	if reader == nil {
		return nil, fmt.Errorf("config: %s input returned a nil reader", kind)
	}
	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("config: read %s input: %w", kind, err)
	}

	return data, nil
}
