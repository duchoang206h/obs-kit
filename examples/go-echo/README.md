# Go Echo Example

Run with the local observability stack:

```bash
make observability-up
make example-go-echo
```

The service listens on `http://localhost:8082`. Send a request:

```bash
curl -i -X POST http://localhost:8082/orders
```

It writes trace-correlated JSON logs to `../../logs/go-echo.log` and exports spans/metrics to the local OpenTelemetry Collector.
