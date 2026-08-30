package grpcclient_test

import (
	"context"
	"net"
	"testing"

	"github.com/devctllabs/go-libs/grpcclient"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
)

func TestNewRequiresTarget(t *testing.T) {
	t.Parallel()

	conn, err := grpcclient.New(grpcclient.Config{})

	require.Nil(t, conn)
	require.ErrorContains(t, err, "target must not be blank")
}

func TestNewConnectsUsingResolverTargetAndCallerInterceptor(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		require.NoError(t, <-serveDone)
	})

	var interceptedMethod string
	conn, err := grpcclient.New(
		grpcclient.Config{Target: "passthrough:///health"},
		grpcclient.WithDialOptions(grpc.WithContextDialer(
			func(context.Context, string) (net.Conn, error) { return listener.Dial() },
		)),
		grpcclient.WithUnaryInterceptors(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			interceptedMethod = method
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	response, err := grpc_health_v1.NewHealthClient(conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})

	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, response.GetStatus())
	require.Equal(t, grpc_health_v1.Health_Check_FullMethodName, interceptedMethod)
}

func TestStreamInterceptorsWrapStreamingCalls(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		require.NoError(t, <-serveDone)
	})

	var interceptedMethod string
	conn, err := grpcclient.New(
		grpcclient.Config{Target: "passthrough:///health"},
		grpcclient.WithDialOptions(grpc.WithContextDialer(
			func(context.Context, string) (net.Conn, error) { return listener.Dial() },
		)),
		grpcclient.WithStreamInterceptors(func(
			ctx context.Context,
			desc *grpc.StreamDesc,
			cc *grpc.ClientConn,
			method string,
			streamer grpc.Streamer,
			opts ...grpc.CallOption,
		) (grpc.ClientStream, error) {
			interceptedMethod = method
			return streamer(ctx, desc, cc, method, opts...)
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	watch, err := grpc_health_v1.NewHealthClient(conn).Watch(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
	_, err = watch.Recv()

	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.Health_Watch_FullMethodName, interceptedMethod)
}

func TestTLSClientCertificateRequiresPrivateKey(t *testing.T) {
	t.Parallel()

	conn, err := grpcclient.New(grpcclient.Config{
		Target: "dns:///service.example",
		TLS: grpcclient.TLSConfig{
			Enabled:         true,
			CertificateFile: "client.pem",
		},
	})

	require.Nil(t, conn)
	require.ErrorContains(t, err, "must be provided together")
}

func TestTLSFilesAreLoadedAtConstruction(t *testing.T) {
	t.Parallel()

	conn, err := grpcclient.New(grpcclient.Config{
		Target: "dns:///service.example",
		TLS: grpcclient.TLSConfig{
			Enabled:    true,
			RootCAFile: "missing-ca.pem",
		},
	})

	require.Nil(t, conn)
	require.ErrorContains(t, err, "read root CA")
}

func TestTelemetryTracesClientCalls(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		require.NoError(t, <-serveDone)
	})

	recorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, tracerProvider.Shutdown(context.Background())) })
	conn, err := grpcclient.New(
		grpcclient.Config{Target: "passthrough:///health"},
		grpcclient.WithDialOptions(grpc.WithContextDialer(
			func(context.Context, string) (net.Conn, error) { return listener.Dial() },
		)),
		grpcclient.WithTelemetry(grpcclient.Telemetry{TracerProvider: tracerProvider}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	_, err = grpc_health_v1.NewHealthClient(conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})

	require.NoError(t, err)
	require.NotEmpty(t, recorder.Ended())
}
