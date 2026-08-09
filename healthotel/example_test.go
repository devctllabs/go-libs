package healthotel_test

import (
	"github.com/devctllabs/go-libs/health"
	"github.com/devctllabs/go-libs/healthotel"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

func ExampleNew() {
	observer, _ := healthotel.New(metricnoop.NewMeterProvider())
	_, _ = health.New(health.WithObserver(observer))
}
