package healthzap_test

import (
	"github.com/devctllabs/go-libs/health"
	"github.com/devctllabs/go-libs/healthzap"
	"go.uber.org/zap"
)

func ExampleNew() {
	observer, _ := healthzap.New(zap.NewNop())
	_, _ = health.New(health.WithObserver(observer))
}
