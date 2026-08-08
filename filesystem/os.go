package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

// OS provides read and mutation access to an operating system directory through
// rooted paths. Create it with Open and close it when no longer needed.
type OS struct {
	root *os.Root
	fsys fs.FS
}

// Open opens root, which must be an existing operating system directory, as a
// rooted filesystem. The caller must close the returned OS.
func Open(root string) (*OS, error) {
	osRoot, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}

	return &OS{root: osRoot, fsys: osRoot.FS()}, nil
}

// Open opens name for reading within the root. Name must satisfy fs.ValidPath.
func (o *OS) Open(name string) (fs.File, error) {
	return o.fsys.Open(name)
}

// Lstat returns information about name without following a final symbolic
// link. Name must satisfy fs.ValidPath.
func (o *OS) Lstat(name string) (fs.FileInfo, error) {
	local, err := localName("lstat", name)
	if err != nil {
		return nil, err
	}

	return o.root.Lstat(local)
}

// ReadLink returns the destination of the symbolic link named name. Name must
// satisfy fs.ValidPath.
func (o *OS) ReadLink(name string) (string, error) {
	local, err := localName("readlink", name)
	if err != nil {
		return "", err
	}

	return o.root.Readlink(local)
}

// Close releases the operating system root. Callers must not use OS after Close.
func (o *OS) Close() error {
	return o.root.Close()
}

func localName(op, name string) (string, error) {
	local, err := filepath.Localize(name)
	if err != nil {
		return "", &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
	}

	return local, nil
}
