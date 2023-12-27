# go-observability-app

Simple golang app that showcases some logging, metrics, tracing and stuff.

The app itself is a super simple GRPC server which demonstrates how to expose OTEL metrics and traces, plus also some log messages which
have the trace info.

The stack includes grafana/loki/tempo/prometheus and shows a bunch of different links between datasources, so that you can easily navigate
between all the different signals.

## Motivation

Why does this repo exist?  Well, a few reasons:

- Many of the capabilities being explored/showcased here are still under active development.  Although grafana documentation is generally
  extremely good, there are parts which are not super clear .. fragmented examples etc.  So, I hope this could prove useful to others as
  a holistic demonstration of how to make it all work together.  There's no attempt here to show how to scale the observability stack itself
  for production workloads, but at least you can see how to expose and correlate the signals.

- The day-job codebase is pretty evolved for a lot of these things but is quite large and currently based on previous SDKs.  Rather than
  using the OTEL SDKs, it uses Prometheus and Jaeger client SDKs because we've been growing it since before even OpenCensus existed. As a
  busy team with a large number of internal users, we're pragmatic about not updating to the latest shiny SDKs "just because", but it seems
  safe to say that OTEL is here to stay – and there are benefits to using it versus sticking with what we have now.  So, this partly exists
  to explore the SDKs, find some nice patterns and see all the current-latest whistles and bells so that we have a clear view of the
  benefits and how to structure things before we migrate.

- Some consolidation of internal services means my team is no longer responsible for running the logs/metrics/traces/grafana services that
  we use – although we do run OTEL Collectors and some Prometheii that remote-write to other places.  So, this also partly exists to ensure
  we have a clear view of all the capabilities offered by the latest incarnations of those services/datasources.

- Lastly, it's a turbulent world out there.  Although I've not been a fully hands-on developer for some time, it's fair to say I'm still an
  active contributor, but much of my work exists on a private github enterprise server.  So, there's no harm in having a public showcase of
  _some_ of my experience.  This is not the only example of that, but since I am something of an observability and service-reliabilty
  zealot, it makes sense to create a body of work that demonstrates that aspect.

## Usage

```plaintext
make docker-build start
make run-client
```

Once running, click through to the following:

- [Boomer Metrics](http://localhost:2223/metrics)
- [Prometheus](http://localhost:9090)
- [Promtail](http://localhost:9080)
- [Grafana](http://localhost:3000)

When tweaking config files, the following gives you a fast clean rebuild:

```plaintext
make stop start
```

Eventually, tear it all down:

```plaintext
make stop
```

Some other make targets also exist, see the following for more details:

```plaintext
make help
```

## TODO

- [ ] Exemplars from application code
- [ ] Additional/improved provisioned dashboards
- [ ] Improve `Logs to Metrics` in loki [datasource](./docker/grafana-datasources.yml)
- [ ] Add github actions for linting and other validation
- [ ] Improve docs – use mkdocs and github pages to provide a rich description of all the capabilities, with screenshots and stuff .. maybe use [snippets](https://facelessuser.github.io/pymdown-extensions/extensions/snippets/)
- [ ] Refactor [server main](./cmd/boomer-server/main.go)
- [ ] Improve [CLI](./cmd/boomer-cli/main.go)
- [ ] [Propagate](https://pkg.go.dev/go.opentelemetry.io/otel@v1.21.0/propagation#TraceContext.Inject) trace context through message queue

Maybe:

- [ ] Add OTEL Collector to show how to add labels if not already set, or always
- [ ] Configure Multi-Tenant
- [ ] [Baggage](https://pkg.go.dev/go.opentelemetry.io/otel@v1.21.0/baggage) and/or use context to define
  shared notion of trace tags and log attributes
- [ ] Unit tests .. [rod](https://go-rod.github.io/#/)?
- [ ] Create Traefik to expose loki/tempo GRPC APIs on default port 9095 .. include some [logcli](https://grafana.com/docs/loki/latest/query/logcli/)/[tempo-cli](https://grafana.com/docs/tempo/latest/operations/tempo_cli/#search) scripts?
- [ ] Use TLS for all endpoints

Future:

- [ ] Open issue/PR on grafana to enable `tracesToMetrics` to use span name
- [ ] ? Add doc section on best practices
