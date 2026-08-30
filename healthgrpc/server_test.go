package healthgrpc_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/health"
	"github.com/devctllabs/go-libs/healthgrpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestNewRejectsNilProbes(t *testing.T) {
	t.Parallel()

	server, err := healthgrpc.New(healthgrpc.Config{}, nil)

	require.Nil(t, server)
	require.ErrorContains(t, err, "Probes must not be nil")
}

func TestCheckEvaluatesTheRequestedProbe(t *testing.T) {
	t.Parallel()

	readinessCalls := 0
	probes, err := health.New(health.Critical("database", health.CheckFunc(func(context.Context) error {
		readinessCalls++
		return errors.New("unavailable")
	})))
	require.NoError(t, err)
	server, err := healthgrpc.New(healthgrpc.Config{}, probes)
	require.NoError(t, err)

	liveness, err := server.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: healthgrpc.LivenessService})
	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, liveness.GetStatus())
	require.Zero(t, readinessCalls)

	for _, service := range []string{"", healthgrpc.ReadinessService} {
		response, err := server.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: service})
		require.NoError(t, err)
		require.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, response.GetStatus())
	}
	require.Equal(t, 2, readinessCalls)

	_, err = server.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: "unknown"})
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestRunPublishesWatchTransitionsAndEndsWatchers(t *testing.T) {
	t.Parallel()

	probes, err := health.New()
	require.NoError(t, err)
	healthServer, err := healthgrpc.New(healthgrpc.Config{PollInterval: time.Hour}, probes)
	require.NoError(t, err)

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	serveDone := make(chan error, 1)
	go func() { serveDone <- grpcServer.Serve(listener) }()
	t.Cleanup(func() {
		grpcServer.Stop()
		require.NoError(t, <-serveDone)
	})
	conn, err := grpc.NewClient(
		"passthrough:///health",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	watch, err := grpc_health_v1.NewHealthClient(conn).Watch(
		context.Background(),
		&grpc_health_v1.HealthCheckRequest{Service: healthgrpc.ReadinessService},
	)
	require.NoError(t, err)
	initial, err := watch.Recv()
	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, initial.GetStatus())
	unknownWatch, err := grpc_health_v1.NewHealthClient(conn).Watch(
		context.Background(),
		&grpc_health_v1.HealthCheckRequest{Service: "unknown"},
	)
	require.NoError(t, err)
	unknownInitial, err := unknownWatch.Recv()
	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN, unknownInitial.GetStatus())

	runCtx, cancelRun := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- healthServer.Run(runCtx) }()
	serving, err := watch.Recv()
	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, serving.GetStatus())

	cancelRun()
	require.NoError(t, <-runDone)
	notServing, err := watch.Recv()
	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, notServing.GetStatus())
	_, err = watch.Recv()
	require.Equal(t, codes.Canceled, status.Code(err))
	unknownDone := make(chan error, 1)
	go func() {
		_, recvErr := unknownWatch.Recv()
		unknownDone <- recvErr
	}()
	select {
	case err := <-unknownDone:
		require.Equal(t, codes.Canceled, status.Code(err))
	case <-time.After(200 * time.Millisecond):
		t.Fatal("unknown service watcher did not end")
	}
}

func TestShutdownIsIdempotentAndPreventsLateProbePublication(t *testing.T) {
	t.Parallel()

	checkStarted := make(chan struct{})
	releaseCheck := make(chan struct{})
	probes, err := health.New(health.Critical("database", health.CheckFunc(func(context.Context) error {
		close(checkStarted)
		<-releaseCheck
		return nil
	})))
	require.NoError(t, err)
	server, err := healthgrpc.New(healthgrpc.Config{PollInterval: time.Hour}, probes)
	require.NoError(t, err)
	runDone := make(chan error, 1)
	go func() { runDone <- server.Run(context.Background()) }()
	<-checkStarted

	require.NoError(t, server.Shutdown(context.Background()))
	require.NoError(t, server.Shutdown(context.Background()))
	close(releaseCheck)
	require.NoError(t, <-runDone)

	listed, err := server.List(context.Background(), &grpc_health_v1.HealthListRequest{})
	require.NoError(t, err)
	for _, response := range listed.GetStatuses() {
		require.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, response.GetStatus())
	}
}
