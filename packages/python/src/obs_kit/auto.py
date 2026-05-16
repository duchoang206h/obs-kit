from __future__ import annotations

from importlib import import_module
from typing import Any


def instrument_python(*, fastapi_app: Any | None = None, flask_app: Any | None = None) -> None:
    try:
        logging_module = import_module("opentelemetry.instrumentation.logging")
        requests_module = import_module("opentelemetry.instrumentation.requests")
    except ImportError as exc:
        raise RuntimeError("Install obs-kit[auto] to use Python auto-instrumentation") from exc

    logging_module.LoggingInstrumentor().instrument(set_logging_format=False)
    requests_module.RequestsInstrumentor().instrument()

    if fastapi_app is not None:
        fastapi_module = import_module("opentelemetry.instrumentation.fastapi")

        fastapi_module.FastAPIInstrumentor.instrument_app(fastapi_app)

    if flask_app is not None:
        flask_module = import_module("opentelemetry.instrumentation.flask")

        flask_module.FlaskInstrumentor().instrument_app(flask_app)
