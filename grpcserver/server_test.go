package grpcserver_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/grpcserver"
	"github.com/devctllabs/go-libs/grpcserver/mocks"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate"
	testvalidatev1 "github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate/v1"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestNewRequiresAddress(t *testing.T) {
	t.Parallel()

	server, err := grpcserver.New(grpcserver.Config{})

	require.Nil(t, server)
	require.ErrorContains(t, err, "address must not be blank")
}

func TestServerRegistersAndServesGeneratedService(t *testing.T) {
	t.Parallel()

	server, err := grpcserver.New(grpcserver.Config{Address: "127.0.0.1:0"})
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:0", server.Address())
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	listener := bufconn.Listen(1024 * 1024)
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	response, err := grpc_health_v1.NewHealthClient(conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, response.GetStatus())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, server.Shutdown(ctx))
	require.NoError(t, <-serveDone)
}

func TestDefaultReflectionCanBeDisabled(t *testing.T) {
	t.Parallel()

	enabled, err := grpcserver.New(grpcserver.Config{Address: "127.0.0.1:0"})
	require.NoError(t, err)
	require.Contains(t, enabled.GetServiceInfo(), "grpc.reflection.v1.ServerReflection")
	require.Contains(t, enabled.GetServiceInfo(), "grpc.reflection.v1alpha.ServerReflection")

	disabled, err := grpcserver.New(grpcserver.Config{
		Address:                  "127.0.0.1:0",
		DisableDefaultReflection: true,
	})
	require.NoError(t, err)
	require.NotContains(t, disabled.GetServiceInfo(), "grpc.reflection.v1.ServerReflection")
	require.NotContains(t, disabled.GetServiceInfo(), "grpc.reflection.v1alpha.ServerReflection")
}

func TestDefaultValidationRejectsInvalidUnaryRequest(t *testing.T) {
	t.Parallel()

	server, err := grpcserver.New(grpcserver.Config{Address: "127.0.0.1:0"})
	require.NoError(t, err)
	testvalidatev1.RegisterTestValidateServiceServer(server, &testvalidate.TestValidateService{})
	conn := serveWithBufconn(t, server)

	_, err = testvalidatev1.NewTestValidateServiceClient(conn).Send(context.Background(), testvalidate.BadUnaryRequest)

	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestRecoveryHidesPanicAndContinuesAfterObserverPanic(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	observer := mocks.NewMockPanicObserver(controller)
	observer.EXPECT().ObservePanic(gomock.Any(), gomock.Any()).Do(func(_ context.Context, event grpcserver.PanicEvent) {
		require.Equal(t, grpc_health_v1.Health_Check_FullMethodName, event.FullMethod)
		require.Equal(t, "secret panic", event.Value)
		require.NotEmpty(t, event.Stack)
	})
	server, err := grpcserver.New(
		grpcserver.Config{Address: "127.0.0.1:0"},
		grpcserver.WithPanicObservers(panickingObserver{}, observer),
	)
	require.NoError(t, err)
	grpc_health_v1.RegisterHealthServer(server, panicHealthServer{})
	conn := serveWithBufconn(t, server)

	_, err = grpc_health_v1.NewHealthClient(conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})

	require.Equal(t, codes.Internal, status.Code(err))
	require.NotContains(t, status.Convert(err).Message(), "secret panic")
}

func TestRecoveryWrapsCallerInterceptors(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	observer := mocks.NewMockPanicObserver(controller)
	observer.EXPECT().ObservePanic(gomock.Any(), gomock.Any())
	server, err := grpcserver.New(
		grpcserver.Config{Address: "127.0.0.1:0"},
		grpcserver.WithPanicObservers(observer),
		grpcserver.WithUnaryInterceptors(func(
			context.Context,
			any,
			*grpc.UnaryServerInfo,
			grpc.UnaryHandler,
		) (any, error) {
			panic("interceptor panic")
		}),
	)
	require.NoError(t, err)
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	conn := serveWithBufconn(t, server)

	_, err = grpc_health_v1.NewHealthClient(conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})

	require.Equal(t, codes.Internal, status.Code(err))
}

func TestRecoveryWrapsCallerStreamInterceptors(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	observer := mocks.NewMockPanicObserver(controller)
	observer.EXPECT().ObservePanic(gomock.Any(), gomock.Any())
	server, err := grpcserver.New(
		grpcserver.Config{Address: "127.0.0.1:0"},
		grpcserver.WithPanicObservers(observer),
		grpcserver.WithStreamInterceptors(func(
			any,
			grpc.ServerStream,
			*grpc.StreamServerInfo,
			grpc.StreamHandler,
		) error {
			panic("stream interceptor panic")
		}),
	)
	require.NoError(t, err)
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	conn := serveWithBufconn(t, server)

	watch, err := grpc_health_v1.NewHealthClient(conn).Watch(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
	_, err = watch.Recv()

	require.Equal(t, codes.Internal, status.Code(err))
}

func TestTLSRequiresServerKeyPair(t *testing.T) {
	t.Parallel()

	server, err := grpcserver.New(grpcserver.Config{
		Address: "127.0.0.1:0",
		TLS:     grpcserver.TLSConfig{Enabled: true},
	})

	require.Nil(t, server)
	require.ErrorContains(t, err, "certificate file must not be blank")
}

func TestTLSFilesAreLoadedAtConstruction(t *testing.T) {
	t.Parallel()

	server, err := grpcserver.New(grpcserver.Config{
		Address: "127.0.0.1:0",
		TLS: grpcserver.TLSConfig{
			Enabled:         true,
			CertificateFile: "missing-cert.pem",
			PrivateKeyFile:  "missing-key.pem",
		},
	})

	require.Nil(t, server)
	require.ErrorContains(t, err, "load TLS key pair")
}

func TestNativeServerOptionsAndRecoveryOptOutAreAccepted(t *testing.T) {
	t.Parallel()

	server, err := grpcserver.New(
		grpcserver.Config{
			Address:                "127.0.0.1:0",
			DisableDefaultRecovery: true,
		},
		grpcserver.WithServerOptions(grpc.MaxRecvMsgSize(1024)),
	)

	require.NoError(t, err)
	require.NotNil(t, server)
}

func TestListenAndServeReportsListenFailure(t *testing.T) {
	t.Parallel()

	server, err := grpcserver.New(grpcserver.Config{Address: "not a valid address"})
	require.NoError(t, err)

	err = server.ListenAndServe()

	require.ErrorContains(t, err, "grpcserver: listen")
}

func TestTelemetryCountsRecoveredPanics(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, meterProvider.Shutdown(context.Background())) })
	server, err := grpcserver.New(
		grpcserver.Config{Address: "127.0.0.1:0"},
		grpcserver.WithTelemetry(grpcserver.Telemetry{MeterProvider: meterProvider}),
	)
	require.NoError(t, err)
	grpc_health_v1.RegisterHealthServer(server, panicHealthServer{})
	conn := serveWithBufconn(t, server)

	_, err = grpc_health_v1.NewHealthClient(conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.Equal(t, codes.Internal, status.Code(err))

	var metrics metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &metrics))
	require.Equal(t, int64(1), panicCount(metrics))
}

func panicCount(metrics metricdata.ResourceMetrics) int64 {
	for _, scope := range metrics.ScopeMetrics {
		for _, measurement := range scope.Metrics {
			if measurement.Name != "grpc.server.panics" {
				continue
			}
			sum, ok := measurement.Data.(metricdata.Sum[int64])
			if ok && len(sum.DataPoints) == 1 {
				return sum.DataPoints[0].Value
			}
		}
	}
	return 0
}

type panickingObserver struct{}

func (panickingObserver) ObservePanic(context.Context, grpcserver.PanicEvent) {
	panic("observer panic")
}

type panicHealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (panicHealthServer) Check(context.Context, *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	panic("secret panic")
}

func serveWithBufconn(t *testing.T, server *grpcserver.Server) *grpc.ClientConn {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.Close())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, server.Shutdown(ctx))
		require.NoError(t, <-serveDone)
	})
	return conn
}
