package health_test

import (
	"context"
	"log"
	"os"

	"github.com/devctllabs/go-libs/health"
)

func Example() {
	logger := log.New(os.Stdout, "", 0)
	probes, _ := health.New(
		health.Critical("database", health.CheckFunc(func(context.Context) error { return nil })),
	)
	logger.Println(probes.Liveness().Status)
	logger.Println(probes.Readiness(context.Background()).Status)

	// Output:
	// ok
	// ok
}
