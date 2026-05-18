# Go HTTP Middleware Example

Run with the Grafana stack:

```bash
make grafana-stack-up
make example-go-http
```

The example starts an HTTP server on `:8080`, creates spans through `HTTPMiddleware`, and logs trace-correlated request data.
