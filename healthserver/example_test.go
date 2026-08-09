package healthserver_test

import (
	"log"
	"os"

	"github.com/devctllabs/go-libs/health"
	"github.com/devctllabs/go-libs/healthserver"
	"github.com/labstack/echo/v5"
)

func ExampleRegister() {
	probes, _ := health.New()
	e := echo.New()
	_ = healthserver.Register(e, probes)
}

func ExampleNewServer() {
	logger := log.New(os.Stdout, "", 0)
	probes, _ := health.New()
	server, _ := healthserver.NewServer(probes)
	logger.Println(server.Address())

	// Output: :8081
}
