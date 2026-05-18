# Obs Kit

Python and Go observability SDKs for personal projects. The project keeps runtime-specific implementations separate while sharing one telemetry contract for configuration, logs, metrics, and traces.

## Packages

- `packages/python/`: Python SDK.
- `packages/go/`: Go SDK.

## Shared Contracts

The `contracts/` directory defines behavior both SDKs should follow:

- environment variable names
- configuration fields
- log attributes
- metric names
- trace attributes

Update contracts before changing behavior that should remain consistent across Python and Go.

## Development

Use the root `Makefile` for common checks:

```bash
make test-python
make test-go
make test
```

The SDKs are intentionally implemented separately. Share contracts and examples, not runtime-specific implementation code.

## Grafana Stack

Start the local Collector, Loki, Tempo, Prometheus, and Grafana stack:

```bash
make grafana-stack-up
```

Run an example, then open Grafana at `http://localhost:3001`.

```bash
make example-python
make example-go
make example-python-otel
make example-python-fastapi
make example-go-echo
```

Use Loki for logs, Tempo for traces, and Prometheus for metrics. New services can use the SDK lifecycle APIs to export traces and metrics over OTLP to the collector at `http://localhost:4318`. Both SDKs include manual DB and Redis operation helpers for client spans plus duration metrics; Python also exposes optional OpenTelemetry auto-instrumentation hooks for FastAPI, Flask, requests, SQLAlchemy, psycopg, asyncpg, Redis, and httpx.

By default, SDK loggers write JSON to stdout/stderr. The examples use `logs/*.log` only so the local collector can tail files during development.

Stop the stack with:

```bash
docker compose -f deployments/grafana-stack/docker-compose.yml down
```

## SigNoz Deployment

Start a SigNoz-backed local stack when you prefer one UI for logs, traces, and metrics:

```bash
make signoz-up
```

The repo compose file at `deployments/signoz/docker-compose.yml` contains SigNoz, ClickHouse, ZooKeeper, and the SigNoz OpenTelemetry Collector. The bridge collector for local file-log examples lives in `deployments/signoz/bridge.docker-compose.yml` and is started by `make signoz-up`.

SigNoz is available at `http://localhost:8080`. Its collector listens on `localhost:4317` and `localhost:4318`. The optional obs-kit bridge collector listens on `localhost:14317` and `localhost:14318`, tails `logs/*.log`, and forwards logs, traces, and metrics to SigNoz. Use the bridge endpoint for local examples:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:14318
```

Pin specific SigNoz image versions with:

```bash
make signoz-up SIGNOZ_VERSION=v0.124.0 SIGNOZ_OTELCOL_VERSION=v0.144.4
```

This deployment is for local development. For any shared environment, set a real `SIGNOZ_TOKENIZER_JWT_SECRET`.

See `deployments/signoz/README.md` for details.

### Grafana Stack Components

The local Docker Compose stack is intended for development and contract testing:

| Component | Local URL | Purpose |
| --- | --- | --- |
| OpenTelemetry Collector | `http://localhost:4318` | Receives OTLP/HTTP telemetry from SDKs. Also exposes OTLP/gRPC on `localhost:4317`. |
| Grafana | `http://localhost:3001` | Explore logs, traces, and metrics through provisioned data sources. Login is disabled locally. |
| Loki | `http://localhost:3100` | Stores structured logs. Grafana data source name: `Loki`. |
| Tempo | `http://localhost:3200` | Stores traces. Grafana data source name: `Tempo`. |
| Prometheus | `http://localhost:9090` | Scrapes metrics exported by the collector. Grafana data source name: `Prometheus`. |

Telemetry flow:

```text
apps/examples
  -> OpenTelemetry Collector (:4318 OTLP/HTTP, :4317 OTLP/gRPC)
  -> Loki for logs
  -> Tempo for traces
  -> Prometheus for metrics
  -> Grafana for exploration
```

The collector configuration is in `deployments/grafana-stack/otel-collector/config.yaml`. It receives OTLP logs, traces, and metrics; tails local example log files from `logs/*.log`; forwards logs to Loki with OTLP/HTTP; forwards traces to Tempo with OTLP/gRPC; and exposes metrics on `otel-collector:8889` for Prometheus to scrape.

### Grafana Explore

Open Grafana at `http://localhost:3001`, then use Explore:

- Logs: select `Loki`, then query by service, for example `{service_name="python-fastapi"}`, `{service_name="python-otel"}`, or `{service_name="go-echo"}`.
- Traces: select `Tempo`, then search recent traces by service name such as `python-fastapi`, `python-otel`, `go-http`, or `go-echo`.
- Metrics: select `Prometheus`, then inspect collector-exported metrics. The local collector exposes SDK metrics to Prometheus through its Prometheus exporter when examples or services emit metric instruments.

For trace-correlated logs, emit a request through one of the HTTP examples, open the related log line in Loki, and use its trace fields to find the matching Tempo trace.

Useful direct endpoints while debugging:

- Collector OTLP/HTTP: `http://localhost:4318`
- Collector OTLP/gRPC: `localhost:4317`
- Loki API: `http://localhost:3100`
- Tempo API: `http://localhost:3200`
- Prometheus UI: `http://localhost:9090`

## Production Use

Use the SDKs as normal dependencies from outside this repository. Pin a tag or commit in production rather than importing local paths.

Python:

```bash
pip install "obs-kit[auto] @ git+https://github.com/duchoang206h/obs-kit.git@<tag-or-commit>#subdirectory=packages/python"
```

Go:

```bash
go get github.com/duchoang206h/obs-kit/packages/go@<tag-or-commit>
```

Configure services with environment variables:

```bash
OTEL_SERVICE_NAME=orders-api
OTEL_SERVICE_VERSION=1.4.2
ENVIRONMENT=production
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_TRACES_SAMPLER_ARG=0.1
OTEL_METRICS_EXPORT_INTERVAL=60000
LOG_LEVEL=info
```

The current SDK exporters use OTLP over HTTP. Set `OTEL_EXPORTER_OTLP_ENDPOINT` to the collector base URL; the SDKs export traces to `/v1/traces` and metrics to `/v1/metrics`.
