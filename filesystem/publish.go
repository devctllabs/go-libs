package filesystem

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
)

const (
	publishFileOperation      = "publishfile"
	publishDirectoryOperation = "publishdirectory"
)

// File describes a regular file to publish. Mode may contain permission bits
// only; publication applies those permissions exactly rather than through the
// process umask.
type File struct {
	Content []byte
	Mode    fs.FileMode
}

// Snapshot describes the complete regular-file content of a directory. Keys
// are paths relative to the published directory and must satisfy fs.ValidPath.
// Directories are inferred from file paths.
type Snapshot map[string]File

// PublishFile replaces name with file through a sibling temporary file. It
// creates missing parent directories and returns false without writing when
// the existing regular file already has the requested content and mode. A
// symbolic-link or non-regular destination returns an error matching
// fs.ErrInvalid.
func (o *OS) PublishFile(ctx context.Context, name string, file File) (bool, error) {
	local, err := publicationName(ctx, publishFileOperation, name)
	if err != nil {
		return false, err
	}
	if err := validateFile(publishFileOperation, name, file); err != nil {
		return false, err
	}

	parent := path.Dir(name)
	localParent, err := localName(publishFileOperation, parent)
	if err != nil {
		return false, err
	}
	if err := o.root.MkdirAll(localParent, 0o777); err != nil {
		return false, err
	}

	equal, err := o.fileEqual(ctx, local, name, file)
	if err != nil {
		return false, err
	}
	if equal {
		return false, nil
	}

	temporary, output, err := o.createTemporaryFile(parent, ".filesystem-publish-file-", file.Mode)
	if err != nil {
		return false, &fs.PathError{Op: publishFileOperation, Path: name, Err: err}
	}
	keepTemporary := true
	defer func() {
		if keepTemporary {
			_ = o.removeAll(temporary)
		}
	}()

	if _, err := output.Write(file.Content); err != nil {
		_ = output.Close()
		return false, &fs.PathError{Op: publishFileOperation, Path: name, Err: err}
	}
	if err := output.Chmod(file.Mode.Perm()); err != nil {
		_ = output.Close()
		return false, &fs.PathError{Op: publishFileOperation, Path: name, Err: err}
	}
	if err := output.Sync(); err != nil {
		_ = output.Close()
		return false, &fs.PathError{Op: publishFileOperation, Path: name, Err: err}
	}
	if err := output.Close(); err != nil {
		return false, &fs.PathError{Op: publishFileOperation, Path: name, Err: err}
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if err := o.root.Rename(temporary, local); err != nil {
		return false, &fs.PathError{Op: publishFileOperation, Path: name, Err: err}
	}
	keepTemporary = false

	return true, nil
}

// PublishDirectory replaces name with the exact contents of snapshot. Files
// are fully prepared in a sibling staging directory before the destination is
// renamed. On systems where sibling renames are atomic this prevents callers
// from observing a partially written tree, but the two-rename replacement can
// briefly leave name absent. A subsequent call recovers an interrupted backup.
// Concurrent publication to the same name is unsupported.
func (o *OS) PublishDirectory(ctx context.Context, name string, snapshot Snapshot) (bool, error) {
	local, err := publicationName(ctx, publishDirectoryOperation, name)
	if err != nil {
		return false, err
	}
	files, err := validateSnapshot(snapshot)
	if err != nil {
		return false, err
	}

	parent := path.Dir(name)
	localParent, err := localName(publishDirectoryOperation, parent)
	if err != nil {
		return false, err
	}
	if err := o.root.MkdirAll(localParent, 0o777); err != nil {
		return false, err
	}

	marker := publicationMarker(name)
	backupName := path.Join(parent, ".filesystem-publish-backup-"+marker)
	backup, err := localName(publishDirectoryOperation, backupName)
	if err != nil {
		return false, err
	}
	stagePrefix := ".filesystem-publish-stage-" + marker + "-"
	if err := o.recoverPublication(ctx, parent, local, backup, stagePrefix); err != nil {
		return false, err
	}

	equal, exists, err := o.snapshotEqual(ctx, name, snapshot)
	if err != nil {
		return false, err
	}
	if equal {
		return false, nil
	}

	stage, err := o.stageSnapshot(ctx, parent, stagePrefix, name, files, snapshot)
	if err != nil {
		return false, err
	}
	keepStage := true
	defer func() {
		if keepStage {
			_ = o.removeAll(stage)
		}
	}()

	changed, err := o.replaceDirectory(local, backup, stage, name, exists)
	if changed {
		keepStage = false
	}
	return changed, err
}

func (o *OS) stageSnapshot(
	ctx context.Context,
	parent, stagePrefix, name string,
	files []string,
	snapshot Snapshot,
) (string, error) {
	stage, err := o.createTemporaryDirectory(parent, stagePrefix)
	if err != nil {
		return "", &fs.PathError{Op: publishDirectoryOperation, Path: name, Err: err}
	}
	keepStage := true
	defer func() {
		if keepStage {
			_ = o.removeAll(stage)
		}
	}()

	for _, fileName := range files {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		file := snapshot[fileName]
		stagedName := path.Join(stage, fileName)
		if err := o.writeStagedFile(stagedName, file); err != nil {
			return "", &fs.PathError{
				Op:   publishDirectoryOperation,
				Path: path.Join(name, fileName),
				Err:  err,
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	keepStage = false
	return stage, nil
}

func (o *OS) replaceDirectory(local, backup, stage, name string, exists bool) (bool, error) {
	if exists {
		if err := o.root.Rename(local, backup); err != nil {
			return false, &fs.PathError{Op: publishDirectoryOperation, Path: name, Err: err}
		}
	}
	if err := o.root.Rename(stage, local); err != nil {
		if exists {
			return false, errors.Join(
				&fs.PathError{Op: publishDirectoryOperation, Path: name, Err: err},
				o.root.Rename(backup, local),
			)
		}
		return false, &fs.PathError{Op: publishDirectoryOperation, Path: name, Err: err}
	}
	if exists {
		if err := o.removeAll(backup); err != nil {
			return true, &fs.PathError{Op: publishDirectoryOperation, Path: name, Err: err}
		}
	}
	return true, nil
}

func publicationName(ctx context.Context, operation, name string) (string, error) {
	local, err := operationName(ctx, operation, name)
	if err != nil {
		return "", err
	}
	if name == "." {
		return "", &fs.PathError{Op: operation, Path: name, Err: fs.ErrInvalid}
	}
	return local, nil
}

func validateFile(operation, name string, file File) error {
	if file.Mode&^fs.ModePerm != 0 {
		return &fs.PathError{Op: operation, Path: name, Err: fs.ErrInvalid}
	}
	return nil
}

func validateSnapshot(snapshot Snapshot) ([]string, error) {
	files := make([]string, 0, len(snapshot))
	for name, file := range snapshot {
		if !fs.ValidPath(name) || name == "." {
			return nil, &fs.PathError{Op: publishDirectoryOperation, Path: name, Err: fs.ErrInvalid}
		}
		if err := validateFile(publishDirectoryOperation, name, file); err != nil {
			return nil, err
		}
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			if _, conflict := snapshot[parent]; conflict {
				return nil, &fs.PathError{Op: publishDirectoryOperation, Path: name, Err: fs.ErrInvalid}
			}
		}
		files = append(files, name)
	}
	sort.Strings(files)
	return files, nil
}

func (o *OS) fileEqual(ctx context.Context, local, name string, expected File) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	info, err := o.root.Lstat(local)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, &fs.PathError{Op: publishFileOperation, Path: name, Err: fs.ErrInvalid}
	}
	if info.Mode().Perm() != expected.Mode.Perm() {
		return false, nil
	}
	content, err := o.root.ReadFile(local)
	if err != nil {
		return false, err
	}
	return bytes.Equal(content, expected.Content), nil
}

func (o *OS) snapshotEqual(ctx context.Context, name string, expected Snapshot) (bool, bool, error) {
	info, err := o.Lstat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	if !info.IsDir() {
		return false, true, &fs.PathError{Op: publishDirectoryOperation, Path: name, Err: fs.ErrInvalid}
	}

	seen := make(map[string]struct{}, len(expected))
	err = fs.WalkDir(o.fsys, name, func(entryName string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		return o.compareSnapshotEntry(ctx, name, entryName, entry, expected, seen)
	})
	if errors.Is(err, errSnapshotDifferent) {
		return false, true, nil
	}
	if err != nil {
		return false, true, err
	}
	return len(seen) == len(expected), true, nil
}

func (o *OS) compareSnapshotEntry(
	ctx context.Context,
	rootName, entryName string,
	entry fs.DirEntry,
	expected Snapshot,
	seen map[string]struct{},
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if entryName == rootName {
		return nil
	}

	relative := strings.TrimPrefix(entryName, rootName+"/")
	if entry.IsDir() {
		if !snapshotHasDirectory(expected, relative) {
			return errSnapshotDifferent
		}
		return nil
	}
	if entry.Type()&fs.ModeType != 0 {
		return errSnapshotDifferent
	}
	file, ok := expected[relative]
	if !ok {
		return errSnapshotDifferent
	}
	info, err := entry.Info()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != file.Mode.Perm() {
		return errSnapshotDifferent
	}
	content, err := fs.ReadFile(o.fsys, entryName)
	if err != nil {
		return err
	}
	if !bytes.Equal(content, file.Content) {
		return errSnapshotDifferent
	}
	seen[relative] = struct{}{}
	return nil
}

var errSnapshotDifferent = errors.New("snapshot differs")

func snapshotHasDirectory(snapshot Snapshot, directory string) bool {
	prefix := directory + "/"
	for name := range snapshot {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func (o *OS) createTemporaryFile(parent, prefix string, mode fs.FileMode) (string, *os.File, error) {
	for range 100 {
		suffix, err := randomSuffix()
		if err != nil {
			return "", nil, err
		}
		name := path.Join(parent, prefix+suffix)
		local, err := localName(publishFileOperation, name)
		if err != nil {
			return "", nil, err
		}
		file, err := o.root.OpenFile(local, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode.Perm())
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		return local, file, err
	}
	return "", nil, fs.ErrExist
}

func (o *OS) createTemporaryDirectory(parent, prefix string) (string, error) {
	for range 100 {
		suffix, err := randomSuffix()
		if err != nil {
			return "", err
		}
		name := path.Join(parent, prefix+suffix)
		local, err := localName(publishDirectoryOperation, name)
		if err != nil {
			return "", err
		}
		err = o.root.Mkdir(local, 0o700)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		return local, err
	}
	return "", fs.ErrExist
}

func randomSuffix() (string, error) {
	var buffer [8]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer[:]), nil
}

func publicationMarker(name string) string {
	sum := sha256.Sum256([]byte(name))
	return hex.EncodeToString(sum[:8])
}

func (o *OS) writeStagedFile(local string, file File) error {
	parent := path.Dir(local)
	if err := o.root.MkdirAll(parent, 0o777); err != nil {
		return err
	}
	output, err := o.root.OpenFile(local, os.O_CREATE|os.O_EXCL|os.O_WRONLY, file.Mode.Perm())
	if err != nil {
		return err
	}
	if _, err := output.Write(file.Content); err != nil {
		_ = output.Close()
		return err
	}
	if err := output.Chmod(file.Mode.Perm()); err != nil {
		_ = output.Close()
		return err
	}
	if err := output.Sync(); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}

func (o *OS) recoverPublication(
	ctx context.Context,
	parent, target, backup, stagePrefix string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	targetExists, err := o.exists(target)
	if err != nil {
		return err
	}
	backupExists, err := o.exists(backup)
	if err != nil {
		return err
	}
	if backupExists && !targetExists {
		if err := o.root.Rename(backup, target); err != nil {
			return err
		}
		backupExists = false
	}
	if backupExists {
		if err := o.removeAll(backup); err != nil {
			return err
		}
	}

	entries, err := fs.ReadDir(o.fsys, parent)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if strings.HasPrefix(entry.Name(), stagePrefix) {
			stage := path.Join(parent, entry.Name())
			localStage, err := localName(publishDirectoryOperation, stage)
			if err != nil {
				return err
			}
			if err := o.removeAll(localStage); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *OS) exists(local string) (bool, error) {
	_, err := o.root.Lstat(local)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

func (o *OS) removeAll(local string) error {
	return o.root.RemoveAll(local)
}
