package filesystem_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/devctllabs/go-libs/filesystem"
	"github.com/stretchr/testify/require"
)

var (
	_ filesystem.Writer  = (*filesystem.OS)(nil)
	_ filesystem.Remover = (*filesystem.OS)(nil)
)

func TestOSWritesAndRemovesRootedEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	ctx := context.Background()
	require.NoError(t, disk.MkdirAll(ctx, "nested/empty", 0o750))
	require.NoError(t, disk.WriteFile(ctx, "nested/file.txt", []byte("content"), 0o600))

	data, err := fs.ReadFile(disk, "nested/file.txt")
	require.NoError(t, err)
	require.Equal(t, "content", string(data))

	require.NoError(t, disk.Remove(ctx, "nested/file.txt"))
	_, err = fs.Stat(disk, "nested/file.txt")
	require.ErrorIs(t, err, fs.ErrNotExist)

	require.NoError(t, disk.Remove(ctx, "nested/empty"))
	_, err = fs.Stat(disk, "nested/empty")
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestOSWriteFileDoesNotCreateParentDirectories(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	err = disk.WriteFile(context.Background(), "missing/file.txt", []byte("content"), 0o600)

	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestOSMutationsHonorCanceledContext(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.ErrorIs(t, disk.MkdirAll(ctx, "directory", 0o750), context.Canceled)
	require.ErrorIs(t, disk.WriteFile(ctx, "file.txt", []byte("content"), 0o600), context.Canceled)
	require.ErrorIs(t, disk.Remove(ctx, "file.txt"), context.Canceled)
	require.NoFileExists(t, filepath.Join(root, "file.txt"))
	require.NoDirExists(t, filepath.Join(root, "directory"))
}

func TestOSMutationsRejectInvalidPaths(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	ctx := context.Background()
	for _, operation := range []func() error{
		func() error { return disk.MkdirAll(ctx, "../directory", 0o750) },
		func() error { return disk.WriteFile(ctx, "/file.txt", nil, 0o600) },
		func() error { return disk.Remove(ctx, "") },
	} {
		err := operation()
		var pathErr *fs.PathError
		require.ErrorAs(t, err, &pathErr)
		require.ErrorIs(t, err, fs.ErrInvalid)
	}
}

func TestOSMutationsCannotEscapeThroughSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require elevated Windows privileges")
	}
	t.Parallel()

	root := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "file.txt")
	require.NoError(t, os.WriteFile(outsideFile, []byte("unchanged"), 0o600))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "escape")))

	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	err = disk.WriteFile(context.Background(), "escape/file.txt", []byte("changed"), 0o600)

	require.Error(t, err)
	data, readErr := os.ReadFile(outsideFile)
	require.NoError(t, readErr)
	require.Equal(t, "unchanged", string(data))
}
