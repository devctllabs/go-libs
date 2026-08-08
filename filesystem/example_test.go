package filesystem_test

import (
	"context"
	"io/fs"
	"log"
	"os"
	"testing/fstest"

	"github.com/devctllabs/go-libs/filesystem"
)

func ExampleOS_Copy() {
	logger := log.New(os.Stdout, "", 0)
	root, err := os.MkdirTemp("", "filesystem-example-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(root) }()

	disk, err := filesystem.Open(root)
	if err != nil {
		log.Print(err)
		return
	}
	defer func() { _ = disk.Close() }()

	source := fstest.MapFS{
		"config/app.toml": {Data: []byte("version = 1")},
	}
	if err := copyPackage(context.Background(), disk, source); err != nil {
		log.Print(err)
		return
	}
	if err := disk.Merge(context.Background(), fstest.MapFS{
		"config/app.toml": {Data: []byte("version = 2")},
	}, "package"); err != nil {
		log.Print(err)
		return
	}

	data, err := fs.ReadFile(disk, "package/config/app.toml")
	if err != nil {
		log.Print(err)
		return
	}
	logger.Println(string(data))

	// Output: version = 2
}

// copyPackage accepts only the filesystem capability it needs. Application
// services would normally depend on a higher-level, consumer-owned interface.
func copyPackage(ctx context.Context, destination filesystem.Copier, source fs.FS) error {
	return destination.Copy(ctx, source, "package")
}
