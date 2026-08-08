package telemetry_test

import (
	"context"
	"log"

	"github.com/devctllabs/go-libs/telemetry"
)

func ExampleOpen() {
	ctx := context.Background()
	runtime, err := telemetry.Open(ctx, telemetry.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := runtime.Shutdown(context.Background()); err != nil {
			log.Printf("shutdown telemetry: %v", err)
		}
	}()

	_ = runtime.TracerProvider()
	_ = runtime.MeterProvider()
	_ = runtime.Propagator()
	// Output:
}
