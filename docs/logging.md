# Logging

The use of `log/slog` gives us easy access to structured logs with no additional dependencies. It allows us to pass around a `*slog.Logger`
that's pre-configured with various log attributes, and a small helper function gets all the details of log/trace/metric linking out of the
way so that application code can focus on logging useful messages.

## Capturing logs

[Promtail](https://grafana.com/docs/loki/latest/send-data/promtail/) is probably the most commonly-used container log shipper,
and that's what we use here.  However, there are (of course) multiple other ways to do this.  Here's a few of them:

- You could use a [Loki slog handler](https://github.com/samber/slog-loki) that sends directly from the application
- If you're using docker, then the [Loki docker plugin](https://grafana.com/docs/loki/latest/send-data/docker-driver/configuration/) can be
  configured either as the default log driver on your host, or as part of the container instantiation.
- [Vector](https://vector.dev/) is also a very capable component, which can pull logs from
  [docker](https://vector.dev/docs/reference/configuration/sources/docker_logs/),
  [kubernetes](https://vector.dev/docs/reference/configuration/sources/kubernetes_logs/) or various other sources.

However, Promtail works very well and it means you can easily route all logs from third-party services via the same mechanism used for the
application. In this example, we use
[docker_sd_configs](https://grafana.com/docs/loki/latest/send-data/promtail/configuration/#docker_sd_config) to discover containers. There's
nothing too involved with the setup, probably the main thing to highlight is the adding of some labels via
[relabel_configs](https://grafana.com/docs/loki/latest/send-data/promtail/configuration/#relabel_configs) – this means we can search for logs
based on these labels, but without adding any other clutter to the log line itself:

```yaml
--8<-- "config/promtail.yml:relabel_configs"
```

You can see all the source labels available for relabelling like this on the Promtail web interface, <http://localhost:9080/targets>.
Check the full config [here](https://github.com/boyvinall/observability-demo/blob/main/config/promtail.yml).

## Application code

We want to make it easy for application developers to get all the signal correlation handled for them.  In our
[GRPC method](https://pkg.go.dev/github.com/boyvinall/observability-demo/pkg/boomerserver#Server.Boom), we simply call

```go
func (s *server) Boom(ctx context.Context, req *pb.BoomRequest) (*pb.BoomResponse, error) {
  logger := util.LoggerFromContext(ctx)
  logger.Info("boom", "boomer_name", req.GetName())
```

This actually logs something like:

``` { .plaintext .wrap }
time=2023-12-30T10:51:41.680Z level=INFO msg=boom hostname=1984ea724676 service_name=MyBoomerServer trace_id=25bb0819a73da590ee2c533162b4fcfa span_id=3c0b58ad9671b7b3 boomer_name="old dude"
```

## Attributes

Notice above that some of the log attributes got added automatically. Of particular importance here are the `trace_id` and `span_id`
attributes, which are pulled from the request context by the
[util.LoggerFromContext](https://pkg.go.dev/github.com/boyvinall/observability-demo/pkg/util#LoggerFromContext) helper function:

```go
--8<-- "pkg/util/logging.go:logger-from-context"
```

For GRPC, you setup the necessary context when you instantiate the `grpc.Server`:

```go
grpcServer := grpc.NewServer(
  grpc.StatsHandler(otelgrpc.NewServerHandler()),
)
```

See reference: [grpc.NewServer](https://pkg.go.dev/google.golang.org/grpc#NewServer),
[grpc.StatsHandler](https://pkg.go.dev/google.golang.org/grpc#StatsHandler) and
[otelgrpc.NewServerHandler](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc#NewServerHandler).

Note that
[otelgrpc.UnaryServerInterceptor](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc#UnaryServerInterceptor)
and [stream](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc#StreamServerInterceptor) server
interceptors are now deprecated – although, more generally, `grpc` unary/stream interceptors are still supported and can be useful for
passing additional context values (more on that later).

The `hostname` and `service_name` attributes in the log line got added as part of the logger
[instantiation](https://pkg.go.dev/github.com/boyvinall/observability-demo/pkg/util#NewLoggerForResource):

```go
--8<-- "pkg/util/logging.go:new-logger-for-resource"
```

In practice, you might prefer the hostname/servicename there to be added by the Promtail `relabel_configs`, but sometimes it can be useful to
setup _some_ global attributes like that.
