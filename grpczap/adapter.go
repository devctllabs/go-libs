package grpczap

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/devctllabs/go-libs/grpcserver"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Adapter exposes zap-backed gRPC interceptors and panic observation.
type Adapter struct {
	logger *zap.Logger
}

// ObservePanic implements grpcserver.PanicObserver.
func (a *Adapter) ObservePanic(ctx context.Context, event grpcserver.PanicEvent) {
	fields := []zap.Field{
		zap.String("rpc.direction", "server"),
		zap.String("rpc.grpc.full_method", event.FullMethod),
		zap.Any("panic", event.Value),
		zap.String("stack", event.Stack),
	}
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		fields = append(fields,
			zap.String("trace_id", spanContext.TraceID().String()),
			zap.String("span_id", spanContext.SpanID().String()),
		)
	}
	a.logger.Error("gRPC server panic recovered", fields...)
}

// StreamServerInterceptor logs one completion record for each streaming server RPC.
func (a *Adapter) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startedAt := time.Now()
		err := handler(srv, stream)
		a.logCompleted(stream.Context(), "server", info.FullMethod, time.Since(startedAt), err)
		return err
	}
}

// UnaryClientInterceptor logs one completion record for each unary client RPC.
func (a *Adapter) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		startedAt := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		a.logCompleted(ctx, "client", method, time.Since(startedAt), err)
		return err
	}
}

// StreamClientInterceptor logs stream creation failures immediately and a
// successful stream when RecvMsg reaches its terminal status.
func (a *Adapter) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		startedAt := time.Now()
		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			a.logCompleted(ctx, "client", method, time.Since(startedAt), err)
			return nil, err
		}
		return &loggingClientStream{
			ClientStream:       stream,
			adapter:            a,
			ctx:                ctx,
			fullMethod:         method,
			startedAt:          startedAt,
			logOnFirstResponse: desc.ClientStreams && !desc.ServerStreams,
		}, nil
	}
}

type loggingClientStream struct {
	grpc.ClientStream
	adapter            *Adapter
	ctx                context.Context
	fullMethod         string
	startedAt          time.Time
	logOnFirstResponse bool
	logOnce            sync.Once
}

func (s *loggingClientStream) RecvMsg(message any) error {
	err := s.ClientStream.RecvMsg(message)
	if err != nil || s.logOnFirstResponse {
		completionErr := err
		if errors.Is(err, io.EOF) {
			completionErr = nil
		}
		s.logOnce.Do(func() {
			s.adapter.logCompleted(s.ctx, "client", s.fullMethod, time.Since(s.startedAt), completionErr)
		})
	}
	return err
}

// UnaryServerInterceptor logs one completion record for each unary server RPC.
func (a *Adapter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startedAt := time.Now()
		response, err := handler(ctx, req)
		a.logCompleted(ctx, "server", info.FullMethod, time.Since(startedAt), err)
		return response, err
	}
}

func (a *Adapter) logCompleted(ctx context.Context, direction, fullMethod string, duration time.Duration, err error) {
	code := status.Code(err)
	fields := []zap.Field{
		zap.String("rpc.direction", direction),
		zap.String("rpc.grpc.full_method", fullMethod),
		zap.String("rpc.grpc.status_code", code.String()),
		zap.Duration("duration", duration),
	}
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		fields = append(fields,
			zap.String("trace_id", spanContext.TraceID().String()),
			zap.String("span_id", spanContext.SpanID().String()),
		)
	}
	if checked := a.logger.Check(levelForCode(code), "gRPC call completed"); checked != nil {
		checked.Write(fields...)
	}
}

func levelForCode(code codes.Code) zapcore.Level {
	switch code {
	case codes.OK, codes.Canceled:
		return zapcore.DebugLevel
	case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.FailedPrecondition,
		codes.OutOfRange, codes.Unauthenticated, codes.PermissionDenied:
		return zapcore.InfoLevel
	case codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted, codes.Unavailable:
		return zapcore.WarnLevel
	default:
		return zapcore.ErrorLevel
	}
}

// New constructs an Adapter using logger.
func New(logger *zap.Logger) (*Adapter, error) {
	if logger == nil {
		return nil, errors.New("grpczap: logger must not be nil")
	}
	return &Adapter{logger: logger}, nil
}
