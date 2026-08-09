package filesystem_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/devctllabs/go-libs/filesystem"
	"github.com/stretchr/testify/require"
)

func TestOSReadsFromRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "file.txt"), []byte("content"), 0o600))

	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	data, err := fs.ReadFile(disk, "file.txt")
	require.NoError(t, err)
	require.Equal(t, "content", string(data))
}

func TestOpenRequiresExistingRoot(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(filepath.Join(t.TempDir(), "missing"))

	require.Nil(t, disk)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestOSCannotReadAfterClose(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, disk.Close())

	_, err = fs.ReadFile(disk, "file.txt")
	require.Error(t, err)
}
