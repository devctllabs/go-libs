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

var (
	_ filesystem.Copier = (*filesystem.OS)(nil)
	_ fs.ReadLinkFS     = (*filesystem.OS)(nil)
)

func TestOSCopiesFilesystemTree(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	source := fstest.MapFS{
		"DOMAIN.md":          {Data: []byte("domain")},
		"agents/worker.toml": {Data: []byte("agent")},
	}

	require.NoError(t, disk.Copy(context.Background(), source, "packages/sample"))

	domain, err := fs.ReadFile(disk, "packages/sample/DOMAIN.md")
	require.NoError(t, err)
	require.Equal(t, "domain", string(domain))

	agent, err := fs.ReadFile(disk, "packages/sample/agents/worker.toml")
	require.NoError(t, err)
	require.Equal(t, "agent", string(agent))
}

func TestOSCopyRejectsExistingFile(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	ctx := context.Background()
	require.NoError(t, disk.MkdirAll(ctx, "target", 0o750))
	require.NoError(t, disk.WriteFile(ctx, "target/file.txt", []byte("existing"), 0o600))

	err = disk.Copy(ctx, fstest.MapFS{"file.txt": {Data: []byte("source")}}, "target")

	require.ErrorIs(t, err, fs.ErrExist)
	data, readErr := fs.ReadFile(disk, "target/file.txt")
	require.NoError(t, readErr)
	require.Equal(t, "existing", string(data))
}

func TestOSCopyRejectsDirectoryOverFile(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	require.NoError(t, disk.WriteFile(context.Background(), "target", nil, 0o600))
	err = disk.Copy(context.Background(), fstest.MapFS{
		"file.txt": {Data: []byte("source")},
	}, "target")

	require.ErrorIs(t, err, fs.ErrExist)
}

func TestOSCopyHonorsContextAndValidatesDestination(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, disk.Copy(ctx, fstest.MapFS{}, "target"), context.Canceled)

	err = disk.Copy(context.Background(), fstest.MapFS{}, "../target")
	var pathErr *fs.PathError
	require.ErrorAs(t, err, &pathErr)
	require.ErrorIs(t, err, fs.ErrInvalid)
}

func TestOSCopyMatchesStandardPermissions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	disk, err := filesystem.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	source := fstest.MapFS{"tool": {Data: []byte("tool"), Mode: 0o751}}
	require.NoError(t, disk.Copy(context.Background(), source, "actual"))
	require.NoError(t, os.CopyFS(filepath.Join(root, "reference"), source))

	actual, err := os.Stat(filepath.Join(root, "actual", "tool"))
	require.NoError(t, err)
	reference, err := os.Stat(filepath.Join(root, "reference", "tool"))
	require.NoError(t, err)
	require.Equal(t, reference.Mode().Perm(), actual.Mode().Perm())
}

func TestOSCopyHandlesSymlinksLikeCopyFS(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require elevated Windows privileges")
	}
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	source := fstest.MapFS{
		"file.txt": {Data: []byte("content")},
		"link.txt": {Data: []byte("file.txt"), Mode: fs.ModeSymlink},
	}
	require.NoError(t, disk.Copy(context.Background(), source, "target"))

	target, err := fs.ReadLink(disk, "target/link.txt")
	require.NoError(t, err)
	require.Equal(t, "file.txt", target)
}

func TestOSCopyRejectsSpecialFiles(t *testing.T) {
	t.Parallel()

	disk, err := filesystem.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, disk.Close()) })

	err = disk.Copy(context.Background(), fstest.MapFS{
		"pipe": {Mode: fs.ModeNamedPipe},
	}, "target")

	require.ErrorIs(t, err, fs.ErrInvalid)
}
