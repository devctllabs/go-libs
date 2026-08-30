package grpczap_test

import (
	"context"
	"testing"

	"github.com/devctllabs/go-libs/grpcserver"
	"github.com/devctllabs/go-libs/grpczap"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNewRejectsNilLogger(t *testing.T) {
	t.Parallel()

	adapter, err := grpczap.New(nil)

	require.Nil(t, adapter)
	require.ErrorContains(t, err, "logger must not be nil")
}

func TestUnaryServerInterceptorLogsCompletedCall(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zapcore.DebugLevel)
	adapter, err := grpczap.New(zap.New(core))
	require.NoError(t, err)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1},
		SpanID:  trace.SpanID{2},
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

	_, err = adapter.UnaryServerInterceptor()(
		ctx,
		"request body must not be logged",
		&grpc.UnaryServerInfo{FullMethod: "/example.Service/Create"},
		func(context.Context, any) (any, error) {
			require.Empty(t, observed.All())
			return nil, status.Error(codes.InvalidArgument, "invalid")
		},
	)

	require.Equal(t, codes.InvalidArgument, status.Code(err))
	entries := observed.All()
	require.Len(t, entries, 1)
	require.Equal(t, zapcore.InfoLevel, entries[0].Level)
	require.Equal(t, "gRPC call completed", entries[0].Message)
	fields := entries[0].ContextMap()
	require.Equal(t, "server", fields["rpc.direction"])
	require.Equal(t, "/example.Service/Create", fields["rpc.grpc.full_method"])
	require.Equal(t, "InvalidArgument", fields["rpc.grpc.status_code"])
	require.Equal(t, spanContext.TraceID().String(), fields["trace_id"])
	require.Equal(t, spanContext.SpanID().String(), fields["span_id"])
	require.Contains(t, fields, "duration")
	require.NotContains(t, fields, "request")
}

func TestStatusCodesUseFixedLogLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code  codes.Code
		level zapcore.Level
	}{
		{codes.OK, zapcore.DebugLevel},
		{codes.Canceled, zapcore.DebugLevel},
		{codes.InvalidArgument, zapcore.InfoLevel},
		{codes.NotFound, zapcore.InfoLevel},
		{codes.AlreadyExists, zapcore.InfoLevel},
		{codes.FailedPrecondition, zapcore.InfoLevel},
		{codes.OutOfRange, zapcore.InfoLevel},
		{codes.Unauthenticated, zapcore.InfoLevel},
		{codes.PermissionDenied, zapcore.InfoLevel},
		{codes.DeadlineExceeded, zapcore.WarnLevel},
		{codes.ResourceExhausted, zapcore.WarnLevel},
		{codes.Aborted, zapcore.WarnLevel},
		{codes.Unavailable, zapcore.WarnLevel},
		{codes.Unknown, zapcore.ErrorLevel},
		{codes.Unimplemented, zapcore.ErrorLevel},
		{codes.Internal, zapcore.ErrorLevel},
		{codes.DataLoss, zapcore.ErrorLevel},
	}
	for _, test := range tests {
		t.Run(test.code.String(), func(t *testing.T) {
			t.Parallel()

			core, observed := observer.New(zapcore.DebugLevel)
			adapter, err := grpczap.New(zap.New(core))
			require.NoError(t, err)
			callErr := error(nil)
			if test.code != codes.OK {
				callErr = status.Error(test.code, "failed")
			}

			_, err = adapter.UnaryServerInterceptor()(
				context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"},
				func(context.Context, any) (any, error) { return nil, callErr },
			)

			require.ErrorIs(t, err, callErr)
			require.Len(t, observed.All(), 1)
			require.Equal(t, test.level, observed.All()[0].Level)
		})
	}
}

func TestClientAndStreamInterceptorsLogCompletedCalls(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zapcore.DebugLevel)
	adapter, err := grpczap.New(zap.New(core))
	require.NoError(t, err)

	err = adapter.UnaryClientInterceptor()(
		context.Background(), "/service/Unary", nil, nil, nil,
		func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
			return status.Error(codes.Unavailable, "unavailable")
		},
	)
	require.Equal(t, codes.Unavailable, status.Code(err))

	err = adapter.StreamServerInterceptor()(
		nil,
		stubServerStream{ctx: context.Background()},
		&grpc.StreamServerInfo{FullMethod: "/service/ServerStream"},
		func(any, grpc.ServerStream) error { return status.Error(codes.DeadlineExceeded, "deadline") },
	)
	require.Equal(t, codes.DeadlineExceeded, status.Code(err))

	_, err = adapter.StreamClientInterceptor()(
		context.Background(),
		&grpc.StreamDesc{},
		nil,
		"/service/ClientStream",
		func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, ...grpc.CallOption) (grpc.ClientStream, error) {
			return nil, status.Error(codes.Unavailable, "unavailable")
		},
	)
	require.Equal(t, codes.Unavailable, status.Code(err))

	entries := observed.All()
	require.Len(t, entries, 3)
	require.Equal(t, "client", entries[0].ContextMap()["rpc.direction"])
	require.Equal(t, "server", entries[1].ContextMap()["rpc.direction"])
	require.Equal(t, "client", entries[2].ContextMap()["rpc.direction"])
}

func TestObservePanicLogsServerDiagnostics(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zapcore.DebugLevel)
	adapter, err := grpczap.New(zap.New(core))
	require.NoError(t, err)

	adapter.ObservePanic(context.Background(), grpcserver.PanicEvent{
		FullMethod: "/example.Service/Create",
		Value:      "secret panic",
		Stack:      "stack trace",
	})

	entries := observed.All()
	require.Len(t, entries, 1)
	require.Equal(t, zapcore.ErrorLevel, entries[0].Level)
	require.Equal(t, "gRPC server panic recovered", entries[0].Message)
	fields := entries[0].ContextMap()
	require.Equal(t, "server", fields["rpc.direction"])
	require.Equal(t, "/example.Service/Create", fields["rpc.grpc.full_method"])
	require.Equal(t, "secret panic", fields["panic"])
	require.Equal(t, "stack trace", fields["stack"])
}

func TestClientStreamingLogsAfterUnaryResponse(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zapcore.DebugLevel)
	adapter, err := grpczap.New(zap.New(core))
	require.NoError(t, err)
	stream, err := adapter.StreamClientInterceptor()(
		context.Background(),
		&grpc.StreamDesc{ClientStreams: true},
		nil,
		"/service/Upload",
		func(ctx context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
			return stubClientStream{ctx: ctx}, nil
		},
	)
	require.NoError(t, err)
	require.Empty(t, observed.All())

	require.NoError(t, stream.RecvMsg(nil))

	entries := observed.All()
	require.Len(t, entries, 1)
	require.Equal(t, zapcore.DebugLevel, entries[0].Level)
	require.Equal(t, "client", entries[0].ContextMap()["rpc.direction"])
}

type stubServerStream struct {
	ctx context.Context
}

func (stubServerStream) SetHeader(metadata.MD) error  { return nil }
func (stubServerStream) SendHeader(metadata.MD) error { return nil }
func (stubServerStream) SetTrailer(metadata.MD)       {}
func (s stubServerStream) Context() context.Context   { return s.ctx }
func (stubServerStream) SendMsg(any) error            { return nil }
func (stubServerStream) RecvMsg(any) error            { return nil }

type stubClientStream struct {
	ctx context.Context
}

func (stubClientStream) Header() (metadata.MD, error) { return nil, nil }
func (stubClientStream) Trailer() metadata.MD         { return nil }
func (stubClientStream) CloseSend() error             { return nil }
func (s stubClientStream) Context() context.Context   { return s.ctx }
func (stubClientStream) SendMsg(any) error            { return nil }
func (stubClientStream) RecvMsg(any) error            { return nil }
