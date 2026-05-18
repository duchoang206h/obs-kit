from __future__ import annotations

import json
import os
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib import request

import uvicorn
from fastapi import FastAPI, HTTPException, Response, status

from obs_kit import Config, configure_logging, init_observability
from obs_kit.auto import instrument_python

SERVICE_NAME = "python-micro-gateway"
SERVICE_VERSION = "0.1.0"
ENVIRONMENT = "local"

inventory_url = os.getenv("INVENTORY_URL", "http://127.0.0.1:8092")
port = int(os.getenv("GATEWAY_PORT", "8091"))

observability = init_observability(
    Config(service_name=SERVICE_NAME, service_version=SERVICE_VERSION, environment=ENVIRONMENT)
)

log_path = Path(__file__).resolve().parents[2] / "logs" / f"{SERVICE_NAME}.log"
logger = configure_logging(observability.config, log_file=log_path, logger_name=SERVICE_NAME)

@asynccontextmanager
async def lifespan(_: FastAPI) -> AsyncIterator[None]:
    try:
        yield
    finally:
        observability.shutdown()


app = FastAPI(lifespan=lifespan)
instrument_python(fastapi_app=app)


@app.post("/orders", status_code=status.HTTP_201_CREATED)
def create_order(response: Response) -> dict[str, object]:
    order_id = "ord_micro_python"

    with observability.start_span("gateway.validate-order") as span:
        span.set_attribute("order.id", order_id)
        logger.info("order validated", extra={"order.id": order_id})

    with observability.start_client_span(
        "gateway.call-inventory",
        attributes={"http.method": "POST", "peer.service": "python-micro-inventory"},
    ) as span:
        payload = json.dumps({"order_id": order_id, "sku": "sku_observability"}).encode()
        headers = {"content-type": "application/json"}
        observability.inject_headers(headers)

        inventory_request = request.Request(
            f"{inventory_url}/reserve",
            data=payload,
            headers=headers,
            method="POST",
        )
        try:
            with request.urlopen(inventory_request, timeout=5) as inventory_response:
                inventory_payload = json.loads(inventory_response.read().decode())
        except HTTPError as exc:
            span.record_exception(exc)
            span.set_attribute("http.response.status_code", exc.code)
            logger.error(
                "inventory request failed",
                extra={"order.id": order_id, "http.response.status_code": exc.code},
            )
            raise HTTPException(
                status_code=status.HTTP_502_BAD_GATEWAY,
                detail="inventory service returned an error",
            ) from exc
        except URLError as exc:
            span.record_exception(exc)
            logger.error(
                "inventory service unavailable",
                extra={"order.id": order_id, "inventory.url": inventory_url},
            )
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="inventory service is unavailable",
            ) from exc

        span.set_attribute("http.response.status_code", inventory_response.status)
        span.set_attribute("inventory.reserved", bool(inventory_payload["reserved"]))
        logger.info(
            "inventory reserved",
            extra={"order.id": order_id, "inventory.reserved": inventory_payload["reserved"]},
        )

    response.headers["x-service-name"] = SERVICE_NAME
    return {"order_id": order_id, "inventory": inventory_payload}


if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=port, log_level="info")
