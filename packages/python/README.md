# Obs Kit Python

Python SDK for the shared observability contracts.

## Development

```bash
uv sync
uv run pytest
uv run ruff check .
uv run mypy src
```

From the repository root, prefer `make test-python` and `make lint-python`; those commands place the virtual environment and uv cache under `/private/tmp`.

The package provides contract-aligned configuration, structured JSON logging, and OpenTelemetry OTLP trace/metric export.

## OpenTelemetry

Use `init_observability()` to configure OTLP trace and metric export. Install `obs-kit[auto]` for optional FastAPI, Flask, requests, and logging auto-instrumentation helpers.

The Python SDK exports telemetry to the configured collector with OTLP over HTTP:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
```

That base endpoint is expanded to `/v1/traces` and `/v1/metrics`.

FastAPI production shape:

```python
from fastapi import FastAPI
from obs_kit import Config, configure_logging, init_observability
from obs_kit.auto import instrument_python

observability = init_observability(
    Config(
        service_name="orders-api",
        service_version="1.4.2",
        environment="production",
    )
)
logger = configure_logging(observability.config, logger_name="orders-api")

app = FastAPI()
instrument_python(app)
```

Use `TracingConfig(sample_rate=0.0)` or `OTEL_TRACES_SAMPLER_ARG=0` when a service should intentionally disable trace sampling.
