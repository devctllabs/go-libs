package filesystem

import (
	"context"
	"io/fs"
)

// MkdirAll creates name and any missing parents within the root using perm,
// subject to the process umask. Name must satisfy fs.ValidPath. If ctx is
// already canceled, MkdirAll returns its error without changing the filesystem.
func (o *OS) MkdirAll(ctx context.Context, name string, perm fs.FileMode) error {
	local, err := operationName(ctx, "mkdirall", name)
	if err != nil {
		return err
	}

	return o.root.MkdirAll(local, perm)
}

// WriteFile writes data to name within the root, creating or truncating the
// file without creating missing parent directories. Name must satisfy
// fs.ValidPath. If ctx is already canceled, WriteFile returns its error without
// changing the filesystem.
func (o *OS) WriteFile(ctx context.Context, name string, data []byte, perm fs.FileMode) error {
	local, err := operationName(ctx, "writefile", name)
	if err != nil {
		return err
	}

	return o.root.WriteFile(local, data, perm)
}

// Remove removes the file or empty directory named name within the root. Name
// must satisfy fs.ValidPath. If ctx is already canceled, Remove returns its
// error without changing the filesystem.
func (o *OS) Remove(ctx context.Context, name string) error {
	local, err := operationName(ctx, "remove", name)
	if err != nil {
		return err
	}

	return o.root.Remove(local)
}

func operationName(ctx context.Context, op, name string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return localName(op, name)
}
