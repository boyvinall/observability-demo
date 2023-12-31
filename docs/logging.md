# Logging

In this demo, we're using the `log/slog` package to give us easy access to structured logs with no additional dependencies. This allows us
to pass around a `*slog.Logger` that's pre-configured with various log attributes. With that, the application code logging useful stuff
doesn't need to worry about anything else, and all the extra stuff that makes correlating logs and traces/etc doesn't get in the way.

## Capturing logs

## Logging trace/span IDs

Most often, the trace/span IDs are available your `context.Context`. For GRPC, you can easily make that happen by passing the
following options when creating your `grpc.Server`:

```go
grpcServer := grpc.NewServer(
  grpc.StatsHandler(otelgrpc.NewServerHandler()),
)
```

Then, in our [GRPC method](https://pkg.go.dev/github.com/boyvinall/observability-demo/pkg/boomerserver#Server.Boom), we simply call

```go
func (s *server) Boom(ctx context.Context, req *pb.BoomRequest) (*pb.BoomResponse, error) {

  logger := util.LoggerFromContext(ctx)
  logger.Info("boom", "name", req.GetName())
```

Inside that [util.LoggerFromContext](https://pkg.go.dev/github.com/boyvinall/observability-demo/pkg/util#LoggerFromContext) function, we
have:

```go
--8<-- "pkg/util/logging.go:logger-from-context"
```

So when the `logger.Info()` is called, it actually logs:

```plaintext
time=2023-12-30T10:51:41.680Z level=INFO msg=boom hostname=1984ea724676 service_name=MyBoomerServer trace_id=25bb0819a73da590ee2c533162b4fcfa span_id=3c0b58ad9671b7b3 boomer_name="old dude"
```

The `hostname` and `service_name` attributes there are from when we initially
[instantiated](https://pkg.go.dev/github.com/boyvinall/observability-demo/pkg/util#NewLoggerForResource) the logger:

```go
--8<-- "pkg/util/logging.go:new-logger-for-resource"
```
