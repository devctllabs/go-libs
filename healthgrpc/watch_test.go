package healthgrpc

import (
	"context"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/health"
	"github.com/stretchr/testify/require"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

func TestSlowWatcherDoesNotBlockAndReceivesLatestStatus(t *testing.T) {
	t.Parallel()

	probes, err := health.New()
	require.NoError(t, err)
	server, err := New(Config{}, probes)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	stream := &blockingWatchStream{
		ctx:     ctx,
		sent:    make(chan grpc_health_v1.HealthCheckResponse_ServingStatus, 2),
		release: make(chan struct{}, 2),
	}
	watchDone := make(chan error, 1)
	go func() {
		watchDone <- server.Watch(
			&grpc_health_v1.HealthCheckRequest{Service: ReadinessService},
			stream,
		)
	}()
	require.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, <-stream.sent)

	updatesDone := make(chan struct{})
	go func() {
		server.setStatus(ReadinessService, grpc_health_v1.HealthCheckResponse_SERVING)
		server.setStatus(ReadinessService, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		server.setStatus(ReadinessService, grpc_health_v1.HealthCheckResponse_SERVING)
		close(updatesDone)
	}()
	select {
	case <-updatesDone:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("status publication blocked on a slow watcher")
	}

	stream.release <- struct{}{}
	require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, <-stream.sent)
	cancel()
	stream.release <- struct{}{}
	require.Error(t, <-watchDone)
}

func TestWatchStartedAfterShutdownReturnsInitialNotServingAndEnds(t *testing.T) {
	t.Parallel()

	probes, err := health.New()
	require.NoError(t, err)
	server, err := New(Config{}, probes)
	require.NoError(t, err)
	require.NoError(t, server.Shutdown(context.Background()))
	stream := &recordingWatchStream{
		ctx:  context.Background(),
		sent: make(chan grpc_health_v1.HealthCheckResponse_ServingStatus, 1),
	}
	watchDone := make(chan error, 1)
	go func() {
		watchDone <- server.Watch(
			&grpc_health_v1.HealthCheckRequest{Service: ReadinessService},
			stream,
		)
	}()

	require.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, <-stream.sent)
	select {
	case err := <-watchDone:
		require.Error(t, err)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("watch started after shutdown did not end")
	}
}

type blockingWatchStream struct {
	ctx     context.Context
	sent    chan grpc_health_v1.HealthCheckResponse_ServingStatus
	release chan struct{}
}

func (s *blockingWatchStream) Send(response *grpc_health_v1.HealthCheckResponse) error {
	s.sent <- response.GetStatus()
	select {
	case <-s.release:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (*blockingWatchStream) SetHeader(metadata.MD) error  { return nil }
func (*blockingWatchStream) SendHeader(metadata.MD) error { return nil }
func (*blockingWatchStream) SetTrailer(metadata.MD)       {}
func (s *blockingWatchStream) Context() context.Context   { return s.ctx }
func (*blockingWatchStream) SendMsg(any) error            { return nil }
func (*blockingWatchStream) RecvMsg(any) error            { return nil }

type recordingWatchStream struct {
	ctx  context.Context
	sent chan grpc_health_v1.HealthCheckResponse_ServingStatus
}

func (s *recordingWatchStream) Send(response *grpc_health_v1.HealthCheckResponse) error {
	s.sent <- response.GetStatus()
	return nil
}

func (*recordingWatchStream) SetHeader(metadata.MD) error  { return nil }
func (*recordingWatchStream) SendHeader(metadata.MD) error { return nil }
func (*recordingWatchStream) SetTrailer(metadata.MD)       {}
func (s *recordingWatchStream) Context() context.Context   { return s.ctx }
func (*recordingWatchStream) SendMsg(any) error            { return nil }
func (*recordingWatchStream) RecvMsg(any) error            { return nil }
