# Python Microservices Example

This example runs two FastAPI services and keeps one distributed trace across both:

- `python-micro-gateway`: receives `POST /orders`, creates gateway spans, and calls inventory.
- `python-micro-inventory`: receives `POST /reserve` and creates an inventory span.

Start the local observability stack:

```bash
make observability-up
```

Start inventory in one terminal:

```bash
make example-python-microservices-inventory
```

Start gateway in another terminal:

```bash
make example-python-microservices-gateway
```

Send a request:

```bash
curl -i -X POST http://localhost:8091/orders
```

View logs in Grafana at `http://localhost:3001` with Loki:

```logql
{service_name="python-micro-gateway"} | log_file_name=~"python-micro-.+\\.log"
```

View the distributed trace in Grafana Explore with Tempo. Search for service name
`python-micro-gateway` or `python-micro-inventory`; a successful request should include spans named:

- `POST /orders`
- `gateway.validate-order`
- `gateway.call-inventory`
- `POST /reserve`
- `inventory.reserve-stock`
