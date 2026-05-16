from pathlib import Path
from contextlib import asynccontextmanager
from collections.abc import AsyncIterator

import uvicorn
from fastapi import FastAPI, Response, status

from obs_kit import Config, configure_logging, init_observability
from obs_kit.auto import instrument_python

observability = init_observability(
    Config(service_name="python-fastapi", service_version="0.1.0", environment="local")
)
log_path = Path(__file__).resolve().parents[2] / "logs" / "python-fastapi.log"
logger = configure_logging(observability.config, log_file=log_path, logger_name="python-fastapi")


@asynccontextmanager
async def lifespan(_: FastAPI) -> AsyncIterator[None]:
    try:
        yield
    finally:
        observability.shutdown()


app = FastAPI(lifespan=lifespan)
instrument_python(fastapi_app=app)


@app.post("/orders", status_code=status.HTTP_201_CREATED)
def create_order(response: Response) -> dict[str, str]:
    logger.info("order created", extra={"order.id": "ord_fastapi_python"})
    response.headers["x-service-name"] = observability.config.service_name or "python-fastapi"
    return {"order_id": "ord_fastapi_python"}


if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=8081, log_level="info")
