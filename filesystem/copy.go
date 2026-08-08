package filesystem

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path"
)

// Copy copies source into destination without overwriting existing files or
// symbolic links. Conflicts return an error matching fs.ErrExist; a nil source
// or invalid destination returns an error matching fs.ErrInvalid. Copy honors
// ctx but is not atomic and may leave a partial result on failure.
func (o *OS) Copy(ctx context.Context, source fs.FS, destination string) error {
	return o.copyTree(ctx, source, destination, "copy", false)
}

// Merge merges source into destination, replacing compatible files and
// symbolic links while preserving directories and unrelated entries. Type
// conflicts return an error matching fs.ErrExist; a nil source or invalid
// destination returns an error matching fs.ErrInvalid. Merge honors ctx but is
// not atomic and may leave a partial result on failure.
func (o *OS) Merge(ctx context.Context, source fs.FS, destination string) error {
	return o.copyTree(ctx, source, destination, "merge", true)
}

func (o *OS) copyTree(ctx context.Context, source fs.FS, destination, operation string, merge bool) error {
	if _, err := operationName(ctx, operation, destination); err != nil {
		return err
	}
	if source == nil {
		return &fs.PathError{Op: operation, Path: ".", Err: fs.ErrInvalid}
	}

	copy := treeCopy{o: o, ctx: ctx, source: source, destination: destination, operation: operation, merge: merge}
	return fs.WalkDir(source, ".", copy.entry)
}

type treeCopy struct {
	o           *OS
	ctx         context.Context
	source      fs.FS
	destination string
	operation   string
	merge       bool
}

func (c treeCopy) entry(sourceName string, entry fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}
	if err := c.ctx.Err(); err != nil {
		return err
	}
	destinationName := path.Join(c.destination, sourceName)
	localDestination, err := localName(c.operation, destinationName)
	if err != nil {
		return err
	}
	switch entry.Type() {
	case fs.ModeDir:
		return c.copyDirectory(localDestination, destinationName)
	case fs.ModeSymlink:
		return c.copySymlink(sourceName, localDestination, destinationName)
	case 0:
		return c.o.copyFile(c.ctx, c.source, sourceName, destinationName, c.operation, c.merge)
	default:
		return &fs.PathError{Op: c.operation, Path: sourceName, Err: fs.ErrInvalid}
	}
}

func (c treeCopy) copyDirectory(localDestination, destinationName string) error {
	mode, exists, err := c.o.destinationMode(localDestination)
	if err != nil {
		return err
	}
	if exists {
		if !mode.IsDir() {
			return conflict(c.operation, destinationName)
		}
		return nil
	}
	return c.o.root.MkdirAll(localDestination, 0o777)
}

func (c treeCopy) copySymlink(sourceName, localDestination, destinationName string) error {
	target, err := fs.ReadLink(c.source, sourceName)
	if err != nil {
		return err
	}
	if err := c.ctx.Err(); err != nil {
		return err
	}
	mode, exists, err := c.o.destinationMode(localDestination)
	if err != nil {
		return err
	}
	if exists {
		if !c.merge || mode&fs.ModeSymlink == 0 {
			return conflict(c.operation, destinationName)
		}
		if err := c.o.root.Remove(localDestination); err != nil {
			return err
		}
	}
	return c.o.root.Symlink(target, localDestination)
}

func (o *OS) copyFile(
	ctx context.Context,
	source fs.FS,
	sourceName, destinationName, operation string,
	merge bool,
) error {
	input, err := source.Open(sourceName)
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()

	info, err := input.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return &fs.PathError{Op: operation, Path: sourceName, Err: fs.ErrInvalid}
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	localDestination, err := localName(operation, destinationName)
	if err != nil {
		return err
	}
	destinationMode, exists, err := o.destinationMode(localDestination)
	if err != nil {
		return err
	}
	if exists && (!merge || !destinationMode.IsRegular()) {
		return conflict(operation, destinationName)
	}

	flags := os.O_CREATE | os.O_EXCL | os.O_WRONLY
	if exists {
		flags = os.O_TRUNC | os.O_WRONLY
	}
	output, err := o.root.OpenFile(
		localDestination,
		flags,
		0o666|info.Mode().Perm()&0o111,
	)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(output, contextReader{ctx: ctx, reader: input})
	closeErr := output.Close()
	if copyErr != nil {
		return &fs.PathError{Op: operation, Path: destinationName, Err: copyErr}
	}
	if closeErr != nil {
		return &fs.PathError{Op: operation, Path: destinationName, Err: closeErr}
	}

	return nil
}

func (o *OS) destinationMode(localName string) (fs.FileMode, bool, error) {
	info, err := o.root.Lstat(localName)
	if err == nil {
		return info.Mode(), true, nil
	}
	if os.IsNotExist(err) {
		return 0, false, nil
	}

	return 0, false, err
}

func conflict(operation, name string) error {
	return &fs.PathError{Op: operation, Path: name, Err: fs.ErrExist}
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}

	return r.reader.Read(buffer)
}
