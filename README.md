# observability-demo

Simple golang demo app to generate logs/metrics/traces, and some infra components to make these signals available in grafana.

The demo consists of:

- Some golang executables:
  - GRPC server
  - worker service which receives messages from the GRPC server over NATS
  - command-line app to post requests to the GRPC server
- Various infra services
  - Grafana - including provisioned datasources and dashboards
  - Loki/Promtail
  - Tempo
  - Prometheus
  - (also a NATS server)

The Grafana datasources are configured with a bunch of different links showing how to navigate between all the different signals.

## Motivation

Why does this repo exist?  Well, a few reasons:

- Many of the capabilities being explored/showcased here are still under active development.  Although Grafana documentation is generally
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

- Lastly, it's a turbulent world out there.  Although I'm not a fully hands-on developer these days, it's fair to say I'm still a very
  active contributor, but much of my work exists on a private github enterprise server.  So, this is a small public example showing
  _some_ of my experience, especially as something of an observability and service-reliabilty zealot.

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

## Docs

The [documentation](./docs/) is built with [mkdocs](https://www.mkdocs.org/) and [mkdocs-material](https://squidfunk.github.io/mkdocs-material/).
It's published as <https://boyvinall.github.io/observability-demo/> but can be served locally as follows:

```plaintext
pip3 install -r requirements.txt
mkdocs serve
```

The code is also reasonably-well documented using syntax described by <https://go.dev/doc/comment>, see
[generated docs](https://pkg.go.dev/github.com/boyvinall/observability-demo). You can run your own doc server
to view local changes as follows:

```plaintext
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite
```

## TODO

- [ ] Improve [docs](./docs/)
- [ ] Improve [CLI](./cmd/boomer-cli/main.go)
- [ ] Additional/improved provisioned dashboards
- [ ] Improve `Logs to Metrics` in loki [datasource](./config/datasources/datasources.yml)
- [ ] Show usage of influxdb as an event logger, including grafana data links
- [ ] Investigate how to support exemplars from application code – currently it seems the golang SDK doesn't support this, see
  [Support exemplars in Prometheus exporter](https://github.com/open-telemetry/opentelemetry-go/issues/3163) and
  [Add support for exemplars](https://github.com/open-telemetry/opentelemetry-go/issues/559). However, there is an
  [OTEP describing integration of exemplars with metrics sdk](https://github.com/open-telemetry/oteps/pull/113). Can also investigate
  how tempo pushes generated metrics.

Maybe:

- [ ] Add OTEL Collector to show how to add labels if not already set, or always
- [ ] Configure Multi-Tenant
- [ ] [Baggage](https://pkg.go.dev/go.opentelemetry.io/otel@v1.21.0/baggage) and/or use context to define
  shared notion of trace tags and log attributes
- [ ] Unit tests .. [rod](https://go-rod.github.io/#/)?
- [ ] Create Traefik to expose loki/tempo GRPC APIs on default port 9095 .. include some [logcli](https://grafana.com/docs/loki/latest/query/logcli/)/[tempo-cli](https://grafana.com/docs/tempo/latest/operations/tempo_cli/#search) scripts?
- [ ] Multiple docker-compose files highlighting different infra topologies
- [ ] Prometheus remote-write tuning

Future:

- [ ] Open issue/PR on grafana to enable `tracesToMetrics` to use span name
- [ ] ? Add doc section on best practices
