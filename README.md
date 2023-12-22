# go-observability-app

Simple golang app that showcases some logging, metrics, tracing and stuff.

The app itself is a super simple GRPC server which demonstrates how to expose OTEL metrics and traces, plus also some log messages which
have the trace info.

The stack includes grafana/loki/tempo/prometheus and shows a bunch of different links between datasources, so that you can easily navigate
between all the different signals.

## Usage

```plaintext
make docker-build start
make run-client
```

Once running, click through to the following:

- [Boomer Metrics](http://localhost:2223/metrics)
- [Prometheus](http://localhost:9090)
- [Grafana](http://localhost:3000)

And eventually tear it all down:

```plaintext
make stop
```

Some other make targets also exist, see the following for more details:

```plaintext
make help
```

## TODO

- [ ] Exemplars
- [ ] Provision some dashboards
- [ ] Improve `Logs to Metrics` in loki [datasource](./docker/grafana-datasources.yml)
- [ ] Expose some metrics directly from the app
- [ ] Add github actions for linting and other validation
- [ ] Improve CLI
- [ ] Improve [readme](./README.md) – describe all the capabilities and add screenshots
- [ ] Refactor [main.go](./cmd/boomer-server/main.go)

Maybe:

- [ ] Create Traefik to enable loki/tempo to both be accessible on default port 9095
- [ ] Multi-Tenant
- [ ] Add OTEL Collector to show how to add labels if not already set, or always
- [ ] Use TLS for all endpoints
- [ ] Unit tests?
- [ ] [Baggage](https://pkg.go.dev/go.opentelemetry.io/otel@v1.21.0/baggage) and/or use context to define
  shared notion of trace tags and log attributes

Future:

- [ ] Open issue/PR on grafana to enable `tracesToMetrics` to use span name
