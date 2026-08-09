package debugserver_test

import (
	"log"
	"os"

	"github.com/devctllabs/go-libs/debugserver"
)

func ExampleNewServer() {
	logger := log.New(os.Stdout, "", 0)
	server, _ := debugserver.NewServer()
	logger.Println(server.Address())

	// Output: 127.0.0.1:6060
}
