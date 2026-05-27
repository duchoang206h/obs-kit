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

The package provides contract-aligned configuration, structured JSON logging, OpenTelemetry OTLP trace/metric export, metric helpers, and optional client instrumentation.

## OpenTelemetry

Use `init_observability()` to configure OTLP trace and metric export. Install `duchoang-obs-kit[auto]` for optional FastAPI, Flask, requests, and logging auto-instrumentation helpers. Install `duchoang-obs-kit[db]`, `duchoang-obs-kit[redis]`, `duchoang-obs-kit[httpx]`, or `duchoang-obs-kit[celery]` when enabling those client integrations.

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
instrument_python(fastapi_app=app)
```

Optional client instrumentation:

```python
instrument_python(
    fastapi_app=app,
    sqlalchemy_engine=engine,
    enable_psycopg=True,
    enable_redis=True,
    enable_httpx=True,
)
```

Celery workers should initialize tracing inside each prefork child process:

```python
from celery.signals import worker_process_init
from obs_kit import Config, init_observability
from obs_kit.auto import instrument_python


@worker_process_init.connect(weak=False)
def init_celery_observability(*_args, **_kwargs):
    init_observability(
        Config(
            service_name="orders-worker",
            service_version="1.4.2",
            environment="production",
        )
    )
    instrument_python(enable_celery=True)
```

Install the Celery extra for this hook:

```bash
pip install "duchoang-obs-kit[celery]"
```

Manual DB or Redis timing works with any client library:

```python
from obs_kit import record_db_operation, record_redis_operation

with record_db_operation(
    tracer_name="orders-api",
    meter_name="orders-api",
    system="postgresql",
    operation="SELECT",
    namespace="orders",
):
    cursor.execute("SELECT id FROM orders WHERE id = %s", (order_id,))

with record_redis_operation(
    tracer_name="orders-api",
    meter_name="orders-api",
    command="GET",
):
    redis.get(cache_key)
```

Application code does not need to import OpenTelemetry directly for common spans or header propagation:

```python
with observability.start_client_span(
    "orders.call-inventory",
    attributes={"peer.service": "inventory-api"},
):
    headers = {"content-type": "application/json"}
    observability.inject_headers(headers)
    call_inventory(headers=headers)
```

Use `TracingConfig(sample_rate=0.0)` or `OTEL_TRACES_SAMPLER_ARG=0` when a service should intentionally disable trace sampling.
