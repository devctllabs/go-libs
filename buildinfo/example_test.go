package buildinfo_test

import (
	"log"
	"os"

	"github.com/devctllabs/go-libs/buildinfo"
)

func ExampleRead() {
	info := buildinfo.Read()
	logger := log.New(os.Stdout, "", 0)
	logger.Printf(
		"module=%s version=%s revision=%s go_version=%s",
		info.ModulePath,
		info.Version,
		info.Revision,
		info.GoVersion,
	)
}
