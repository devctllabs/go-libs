package grpcserver

import (
	"context"
	"errors"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate go tool mockgen -destination mocks/panic_observer.gen.go -package mocks . PanicObserver

// PanicEvent describes a panic recovered at an RPC boundary.
type PanicEvent struct {
	FullMethod string
	Value      any
	Stack      string
}

// PanicObserver receives recovered panic diagnostics. Implementations must
// return promptly and must not retain ctx after ObservePanic returns.
type PanicObserver interface {
	ObservePanic(ctx context.Context, event PanicEvent)
}

// WithPanicObservers appends observers notified after a recovered RPC panic.
func WithPanicObservers(observers ...PanicObserver) Option {
	return optionFunc(func(cfg *serverConfig) error {
		for _, observer := range observers {
			if observer == nil {
				return errors.New("grpcserver: panic observer must not be nil")
			}
			cfg.panicObservers = append(cfg.panicObservers, observer)
		}
		return nil
	})
}

func recoveryUnaryInterceptor(observers []PanicObserver) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer recoverRPCPanic(ctx, info.FullMethod, observers, &err)
		return handler(ctx, req)
	}
}

func recoveryStreamInterceptor(observers []PanicObserver) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer recoverRPCPanic(stream.Context(), info.FullMethod, observers, &err)
		return handler(srv, stream)
	}
}

func recoverRPCPanic(ctx context.Context, fullMethod string, observers []PanicObserver, err *error) {
	value := recover()
	if value == nil {
		return
	}
	event := PanicEvent{FullMethod: fullMethod, Value: value, Stack: string(debug.Stack())}
	for _, observer := range observers {
		notifyPanicObserver(ctx, observer, event)
	}
	*err = status.Error(codes.Internal, "internal server error")
}

func notifyPanicObserver(ctx context.Context, observer PanicObserver, event PanicEvent) {
	defer func() { _ = recover() }()
	observer.ObservePanic(ctx, event)
}
