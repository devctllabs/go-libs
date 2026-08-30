# grpczap

`grpczap` is an optional zap adapter for `grpcserver` and `grpcclient`. Its four
interceptors write one completion record per RPC without logging payloads or
metadata. The same adapter implements `grpcserver.PanicObserver`.

```go
adapter, err := grpczap.New(logger)
if err != nil {
	return err
}

server, err := grpcserver.New(
	grpcserver.Config{Address: ":9000"},
	grpcserver.WithUnaryInterceptors(adapter.UnaryServerInterceptor()),
	grpcserver.WithStreamInterceptors(adapter.StreamServerInterceptor()),
	grpcserver.WithPanicObservers(adapter),
)
```

Completion records include direction, full method, status code, duration, and
available trace/span IDs. Only recovered server panics include the recovered
value and stack.
