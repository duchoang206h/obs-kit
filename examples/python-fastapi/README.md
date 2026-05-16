# Python FastAPI Example

Run with the local observability stack:

```bash
make observability-up
make example-python-fastapi
```

The service listens on `http://localhost:8081`. Send a request:

```bash
curl -i -X POST http://localhost:8081/orders
```

It writes trace-correlated JSON logs to `../../logs/python-fastapi.log` and exports spans/metrics to the local OpenTelemetry Collector.

View logs in Grafana at `http://localhost:3001` with the Loki query:

```logql
{service_name="python-fastapi"}
```
