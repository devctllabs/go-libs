package healthserver_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devctllabs/go-libs/health"
	"github.com/devctllabs/go-libs/health/mocks"
	"github.com/devctllabs/go-libs/healthserver"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegisterServesLiveness(t *testing.T) {
	t.Parallel()
	probes, err := health.New()
	require.NoError(t, err)

	e := echo.New()
	require.NoError(t, healthserver.Register(e, probes))

	request := httptest.NewRequest(http.MethodGet, "/livez", nil)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func TestRegisterDoesNotExposeStartupProbe(t *testing.T) {
	t.Parallel()
	probes, err := health.New()
	require.NoError(t, err)
	e := echo.New()
	require.NoError(t, healthserver.Register(e, probes))

	recorder := serve(t, e, "/startupz")
	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestReadinessRunsChecksButOnlyVerboseReturnsThem(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)
	checker.EXPECT().Check(gomock.Any()).Return(nil).Times(2)

	probes, err := health.New(health.Critical("postgres", checker))
	require.NoError(t, err)
	e := echo.New()
	require.NoError(t, healthserver.Register(e, probes))

	brief := serve(t, e, "/readyz")
	require.Equal(t, http.StatusOK, brief.Code)
	require.JSONEq(t, `{"status":"ok"}`, brief.Body.String())

	verbose := serve(t, e, "/readyz?verbose=true")
	require.Equal(t, http.StatusOK, verbose.Code)
	require.JSONEq(t, `{
		"status":"ok",
		"checks":[{"name":"postgres","status":"ok","critical":true}]
	}`, verbose.Body.String())
}

func TestReadinessVerboseFalseReturnsBriefResponse(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)
	checker.EXPECT().Check(gomock.Any()).Return(nil)
	probes, err := health.New(health.Critical("postgres", checker))
	require.NoError(t, err)
	e := echo.New()
	require.NoError(t, healthserver.Register(e, probes))

	recorder := serve(t, e, "/readyz?verbose=false")
	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func TestReadinessCriticalFailureReturnsUnavailableWithoutErrorDetails(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)
	checker.EXPECT().Check(gomock.Any()).Return(errors.New("postgres://secret@db.internal unavailable"))

	probes, err := health.New(health.Critical("postgres", checker))
	require.NoError(t, err)
	e := echo.New()
	require.NoError(t, healthserver.Register(e, probes))

	recorder := serve(t, e, "/readyz?verbose=true")
	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.JSONEq(t, `{
		"status":"fail",
		"checks":[{"name":"postgres","status":"fail","critical":true}]
	}`, recorder.Body.String())
	require.NotContains(t, recorder.Body.String(), "secret")
}

func TestReadinessNonCriticalFailureKeepsOverallStatusOK(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)
	checker.EXPECT().Check(gomock.Any()).Return(errors.New("cache unavailable"))

	probes, err := health.New(health.NonCritical("cache", checker))
	require.NoError(t, err)
	e := echo.New()
	require.NoError(t, healthserver.Register(e, probes))

	recorder := serve(t, e, "/readyz?verbose=true")
	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{
		"status":"ok",
		"checks":[{"name":"cache","status":"fail","critical":false}]
	}`, recorder.Body.String())
}

func TestReadinessRejectsMalformedVerboseAsProblemDetails(t *testing.T) {
	t.Parallel()
	probes, err := health.New()
	require.NoError(t, err)
	e := echo.New()
	require.NoError(t, healthserver.Register(e, probes))

	for _, target := range []string{"/readyz?verbose=invalid", "/readyz?verbose"} {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			recorder := serve(t, e, target)
			require.Equal(t, http.StatusBadRequest, recorder.Code)
			require.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))
			require.JSONEq(t, `{
				"type":"/problems/bad-request",
				"title":"Bad Request",
				"status":400,
				"retryable":false
			}`, recorder.Body.String())
		})
	}
}

func TestStandaloneServerServesAndShutsDown(t *testing.T) {
	t.Parallel()
	probes, err := health.New()
	require.NoError(t, err)
	server, err := healthserver.NewServer(probes)
	require.NoError(t, err)
	require.Equal(t, ":8081", server.Address())

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	runErr := make(chan error, 1)
	go func() {
		runErr <- server.Serve(listener)
	}()

	response, err := http.Get("http://" + listener.Addr().String() + "/livez")
	require.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.JSONEq(t, `{"status":"ok"}`, string(body))

	require.NoError(t, server.Shutdown(context.Background()))
	require.NoError(t, <-runErr)
}

func TestStandaloneServerValidatesInputs(t *testing.T) {
	t.Parallel()
	_, err := healthserver.NewServer(nil)
	require.ErrorContains(t, err, "must not be nil")

	probes, err := health.New()
	require.NoError(t, err)
	_, err = healthserver.NewServer(probes, healthserver.WithAddress(" "))
	require.ErrorContains(t, err, "must not be blank")

	require.Error(t, healthserver.Register(nil, probes))
	require.Error(t, healthserver.Register(echo.New(), nil))
}

func serve(t *testing.T, e *echo.Echo, target string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)
	return recorder
}
