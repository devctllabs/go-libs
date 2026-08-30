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

func TestOSPublishFileCreatesParentsAndSkipsUnchangedContent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	changed, err := disk.PublishFile(context.Background(), "nested/file.txt", filesystem.File{
		Content: []byte("first"),
		Mode:    0o640,
	})
	require.NoError(t, err)
	require.True(t, changed)
	require.FileExists(t, filepath.Join(root, "nested", "file.txt"))
	requireFile(t, root, "nested/file.txt", "first", 0o640)

	changed, err = disk.PublishFile(context.Background(), "nested/file.txt", filesystem.File{
		Content: []byte("first"),
		Mode:    0o640,
	})
	require.NoError(t, err)
	require.False(t, changed)

	changed, err = disk.PublishFile(context.Background(), "nested/file.txt", filesystem.File{
		Content: []byte("second"),
		Mode:    0o600,
	})
	require.NoError(t, err)
	require.True(t, changed)
	requireFile(t, root, "nested/file.txt", "second", 0o600)
}

func TestOSPublishFileRejectsSymlinkTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require elevated Windows privileges")
	}
	t.Parallel()

	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	require.NoError(t, os.WriteFile(outside, []byte("unchanged"), 0o600))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "file.txt")))

	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	_, err = disk.PublishFile(context.Background(), "file.txt", filesystem.File{
		Content: []byte("changed"),
		Mode:    0o600,
	})
	require.ErrorIs(t, err, fs.ErrInvalid)

	content, err := os.ReadFile(outside)
	require.NoError(t, err)
	require.Equal(t, "unchanged", string(content))
}

func TestOSPublishDirectoryReplacesTheWholeSnapshot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "generated", "obsolete"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "generated", "obsolete", "old.txt"), []byte("old"), 0o600))

	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	snapshot := filesystem.Snapshot{
		"README.md":        {Content: []byte("generated\n"), Mode: 0o644},
		"cmd/tool/main.go": {Content: []byte("package main\n"), Mode: 0o600},
	}
	changed, err := disk.PublishDirectory(context.Background(), "generated", snapshot)
	require.NoError(t, err)
	require.True(t, changed)
	require.NoDirExists(t, filepath.Join(root, "generated", "obsolete"))
	requireFile(t, root, "generated/README.md", "generated\n", 0o644)
	requireFile(t, root, "generated/cmd/tool/main.go", "package main\n", 0o600)

	changed, err = disk.PublishDirectory(context.Background(), "generated", snapshot)
	require.NoError(t, err)
	require.False(t, changed)

	changed, err = disk.PublishDirectory(context.Background(), "generated", filesystem.Snapshot{})
	require.NoError(t, err)
	require.True(t, changed)
	entries, err := os.ReadDir(filepath.Join(root, "generated"))
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestOSPublishDirectoryRejectsInvalidSnapshotBeforeMutation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "generated"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "generated", "existing.txt"), []byte("keep"), 0o600))

	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	_, err = disk.PublishDirectory(context.Background(), "generated", filesystem.Snapshot{
		"../escape.txt": {Content: []byte("bad"), Mode: 0o600},
	})
	require.ErrorIs(t, err, fs.ErrInvalid)
	requireFile(t, root, "generated/existing.txt", "keep", 0o600)
}

func TestOSPublishDirectoryRejectsFileAndParentConflict(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	_, err = disk.PublishDirectory(context.Background(), "generated", filesystem.Snapshot{
		"a":        {Content: []byte("file"), Mode: 0o600},
		"a-middle": {Content: []byte("file"), Mode: 0o600},
		"a/b.txt":  {Content: []byte("nested"), Mode: 0o600},
	})
	require.ErrorIs(t, err, fs.ErrInvalid)
}

func TestOSPublicationHonorsCanceledContextBeforeMutation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = disk.PublishFile(ctx, "file.txt", filesystem.File{Content: []byte("content"), Mode: 0o600})
	require.ErrorIs(t, err, context.Canceled)
	_, err = disk.PublishDirectory(ctx, "generated", filesystem.Snapshot{
		"file.txt": {Content: []byte("content"), Mode: 0o600},
	})
	require.ErrorIs(t, err, context.Canceled)
	require.NoFileExists(t, filepath.Join(root, "file.txt"))
	require.NoDirExists(t, filepath.Join(root, "generated"))
}

func TestOSRemoveAllRemovesTreeAndHonorsContext(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "nested", "child"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "nested", "child", "file.txt"), []byte("content"), 0o600))

	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, disk.RemoveAll(canceled, "nested"), context.Canceled)
	require.DirExists(t, filepath.Join(root, "nested"))

	require.NoError(t, disk.RemoveAll(context.Background(), "nested"))
	require.NoDirExists(t, filepath.Join(root, "nested"))
}

func requireFile(t *testing.T, root, name, content string, mode fs.FileMode) {
	t.Helper()

	fullName := filepath.Join(root, filepath.FromSlash(name))
	data, err := os.ReadFile(fullName)
	require.NoError(t, err)
	require.Equal(t, content, string(data))
	info, err := os.Stat(fullName)
	require.NoError(t, err)
	require.Equal(t, mode, info.Mode().Perm())
}
