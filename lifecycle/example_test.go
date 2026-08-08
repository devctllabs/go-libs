package lifecycle_test

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/devctllabs/go-libs/lifecycle"
)

func ExampleRun() {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	go func() {
		<-started
		cancel()
	}()

	err := lifecycle.Run(ctx, lifecycle.Config{
		ShutdownTimeout: 5 * time.Second,
		Shutdown:        func(context.Context) error { return nil },
		Tasks: []lifecycle.Task{{
			Name: "api",
			Run: func(ctx context.Context) error {
				close(started)
				<-ctx.Done()
				return ctx.Err()
			},
		}},
	})
	logger := log.New(os.Stdout, "", 0)
	logger.Println(err)

	// Output: <nil>
}
