package debugserver_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/debugserver"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	t.Parallel()

	t.Run("uses the loopback default address", func(t *testing.T) {
		t.Parallel()

		server, err := debugserver.NewServer()

		require.NoError(t, err)
		require.Equal(t, "127.0.0.1:6060", server.Address())
	})

	t.Run("accepts a custom address", func(t *testing.T) {
		t.Parallel()

		server, err := debugserver.NewServer(debugserver.WithAddress("127.0.0.1:0"))

		require.NoError(t, err)
		require.Equal(t, "127.0.0.1:0", server.Address())
	})

	t.Run("ignores a nil option", func(t *testing.T) {
		t.Parallel()

		server, err := debugserver.NewServer(nil)

		require.NoError(t, err)
		require.Equal(t, "127.0.0.1:6060", server.Address())
	})

	t.Run("rejects a blank address", func(t *testing.T) {
		t.Parallel()

		server, err := debugserver.NewServer(debugserver.WithAddress(" \t "))

		require.EqualError(t, err, "debugserver: address must not be blank")
		require.Nil(t, server)
	})
}

func TestServerServeRejectsNilListener(t *testing.T) {
	t.Parallel()

	server, err := debugserver.NewServer()
	require.NoError(t, err)

	require.EqualError(t, server.Serve(nil), "debugserver: listener must not be nil")
}

func TestServerServesStandardPprofEndpoints(t *testing.T) {
	t.Parallel()
	server, err := debugserver.NewServer()
	require.NoError(t, err)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, server.Shutdown(shutdownCtx))
		require.NoError(t, <-serveErr)
	})

	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://" + listener.Addr().String()

	for _, path := range []string{
		"/debug/pprof/",
		"/debug/pprof/goroutine?debug=1",
		"/debug/pprof/cmdline",
		"/debug/pprof/symbol",
		"/debug/pprof/profile?seconds=1",
		"/debug/pprof/trace?seconds=0.01",
	} {
		response, requestErr := client.Get(baseURL + path)
		require.NoError(t, requestErr, path)
		_, requestErr = io.Copy(io.Discard, response.Body)
		require.NoError(t, requestErr, path)
		require.NoError(t, response.Body.Close(), path)
		require.Equal(t, http.StatusOK, response.StatusCode, path)
	}

	response, err := client.Get(baseURL + "/debug/pprof/unknown")
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusNotFound, response.StatusCode)

	response, err = client.Get(baseURL + "/")
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusNotFound, response.StatusCode)

	request, err := http.NewRequest(http.MethodPost, baseURL+"/debug/pprof/", nil)
	require.NoError(t, err)
	response, err = client.Do(request)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
}
