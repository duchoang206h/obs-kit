# Go HTTP Middleware Example

Run with the local stack:

```bash
make observability-up
make example-go-http
```

The example starts an HTTP server on `:8080`, creates spans through `HTTPMiddleware`, and logs trace-correlated request data.
