from __future__ import annotations

import os
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from pathlib import Path

import uvicorn
from fastapi import FastAPI

from obs_kit import Config, configure_logging, init_observability
from obs_kit.auto import instrument_python

SERVICE_NAME = "python-micro-inventory"
SERVICE_VERSION = "0.1.0"
ENVIRONMENT = "local"

port = int(os.getenv("INVENTORY_PORT", "8092"))

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


@app.post("/reserve")
def reserve_inventory(payload: dict[str, str]) -> dict[str, object]:
    order_id = payload["order_id"]
    sku = payload["sku"]

    with observability.start_span("inventory.reserve-stock") as span:
        span.set_attribute("order.id", order_id)
        span.set_attribute("inventory.sku", sku)
        logger.info("stock reserved", extra={"order.id": order_id, "inventory.sku": sku})

    return {"order_id": order_id, "sku": sku, "reserved": True}


if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=port, log_level="info")
