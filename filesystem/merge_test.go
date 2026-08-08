package filesystem_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/devctllabs/go-libs/filesystem"
	"github.com/stretchr/testify/require"
)

var _ filesystem.Merger = (*filesystem.OS)(nil)

func TestOSMergesCompatibleFilesystemTree(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	ctx := context.Background()
	require.NoError(t, disk.MkdirAll(ctx, "target", 0o750))
	require.NoError(t, disk.WriteFile(ctx, "target/file.txt", []byte("old"), 0o600))
	require.NoError(t, disk.WriteFile(ctx, "target/extra.txt", []byte("extra"), 0o600))

	source := fstest.MapFS{
		"file.txt":       {Data: []byte("new"), Mode: 0o755},
		"nested/new.txt": {Data: []byte("nested")},
	}
	require.NoError(t, disk.Merge(ctx, source, "target"))

	updated, err := fs.ReadFile(disk, "target/file.txt")
	require.NoError(t, err)
	require.Equal(t, "new", string(updated))
	extra, err := fs.ReadFile(disk, "target/extra.txt")
	require.NoError(t, err)
	require.Equal(t, "extra", string(extra))
	nested, err := fs.ReadFile(disk, "target/nested/new.txt")
	require.NoError(t, err)
	require.Equal(t, "nested", string(nested))

	info, err := os.Stat(filepath.Join(root, "target", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, fs.FileMode(0o600), info.Mode().Perm())
}

func TestOSMergeRejectsTypeMismatches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		source      fstest.MapFS
		preparePath func(t *testing.T, disk *filesystem.OS)
	}{
		{
			name:   "file over directory",
			source: fstest.MapFS{"entry": {Data: []byte("file")}},
			preparePath: func(t *testing.T, disk *filesystem.OS) {
				require.NoError(t, disk.MkdirAll(context.Background(), "target/entry", 0o750))
			},
		},
		{
			name:   "directory over file",
			source: fstest.MapFS{"entry": {Mode: fs.ModeDir}},
			preparePath: func(t *testing.T, disk *filesystem.OS) {
				require.NoError(t, disk.MkdirAll(context.Background(), "target", 0o750))
				require.NoError(t, disk.WriteFile(context.Background(), "target/entry", nil, 0o600))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			disk, err := filesystem.Open(t.TempDir())
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, disk.Close()) })
			tt.preparePath(t, disk)

			err = disk.Merge(context.Background(), tt.source, "target")

			require.ErrorIs(t, err, fs.ErrExist)
		})
	}
}

func TestOSMergeReplacesSymlinkTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require elevated Windows privileges")
	}
	t.Parallel()

	root := t.TempDir()
	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	require.NoError(t, disk.MkdirAll(context.Background(), "target", 0o750))
	require.NoError(t, os.Symlink("old.txt", filepath.Join(root, "target", "link.txt")))
	source := fstest.MapFS{
		"link.txt": {Data: []byte("new.txt"), Mode: fs.ModeSymlink},
	}

	require.NoError(t, disk.Merge(context.Background(), source, "target"))
	target, err := fs.ReadLink(disk, "target/link.txt")
	require.NoError(t, err)
	require.Equal(t, "new.txt", target)
}

func TestOSMergeRejectsSymlinkTypeMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require elevated Windows privileges")
	}
	t.Parallel()

	root := t.TempDir()
	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	require.NoError(t, disk.MkdirAll(context.Background(), "target", 0o750))
	require.NoError(t, disk.WriteFile(context.Background(), "target/link.txt", nil, 0o600))
	source := fstest.MapFS{
		"link.txt": {Data: []byte("file.txt"), Mode: fs.ModeSymlink},
	}

	err = disk.Merge(context.Background(), source, "target")

	require.ErrorIs(t, err, fs.ErrExist)
}

func TestOSMergeRejectsFileOverSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require elevated Windows privileges")
	}
	t.Parallel()

	root := t.TempDir()
	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	require.NoError(t, disk.MkdirAll(context.Background(), "target", 0o750))
	require.NoError(t, os.Symlink("other.txt", filepath.Join(root, "target", "file.txt")))

	err = disk.Merge(context.Background(), fstest.MapFS{
		"file.txt": {Data: []byte("source")},
	}, "target")

	require.ErrorIs(t, err, fs.ErrExist)
}
