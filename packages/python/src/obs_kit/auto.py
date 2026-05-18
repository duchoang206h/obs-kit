from __future__ import annotations

from importlib import import_module
from typing import Any


def instrument_python(
    *,
    fastapi_app: Any | None = None,
    flask_app: Any | None = None,
    sqlalchemy_engine: Any | None = None,
    enable_asyncpg: bool = False,
    enable_httpx: bool = False,
    enable_psycopg: bool = False,
    enable_psycopg2: bool = False,
    enable_redis: bool = False,
) -> None:
    logging_module = _import_optional(
        "opentelemetry.instrumentation.logging",
        "Install obs-kit[auto] to use logging auto-instrumentation",
    )
    requests_module = _import_optional(
        "opentelemetry.instrumentation.requests",
        "Install obs-kit[auto] to use requests auto-instrumentation",
    )
    logging_module.LoggingInstrumentor().instrument(set_logging_format=False)
    requests_module.RequestsInstrumentor().instrument()

    if fastapi_app is not None:
        fastapi_module = _import_optional(
            "opentelemetry.instrumentation.fastapi",
            "Install obs-kit[auto] to instrument FastAPI",
        )
        fastapi_module.FastAPIInstrumentor.instrument_app(fastapi_app)

    if flask_app is not None:
        flask_module = _import_optional(
            "opentelemetry.instrumentation.flask",
            "Install obs-kit[auto] to instrument Flask",
        )
        flask_module.FlaskInstrumentor().instrument_app(flask_app)

    if sqlalchemy_engine is not None:
        sqlalchemy_module = _import_optional(
            "opentelemetry.instrumentation.sqlalchemy",
            "Install obs-kit[db] to instrument SQLAlchemy",
        )
        sqlalchemy_module.SQLAlchemyInstrumentor().instrument(engine=sqlalchemy_engine)

    if enable_asyncpg:
        asyncpg_module = _import_optional(
            "opentelemetry.instrumentation.asyncpg",
            "Install obs-kit[db] to instrument asyncpg",
        )
        asyncpg_module.AsyncPGInstrumentor().instrument()

    if enable_psycopg:
        psycopg_module = _import_optional(
            "opentelemetry.instrumentation.psycopg",
            "Install obs-kit[db] to instrument psycopg",
        )
        psycopg_module.PsycopgInstrumentor().instrument()

    if enable_psycopg2:
        psycopg2_module = _import_optional(
            "opentelemetry.instrumentation.psycopg2",
            "Install obs-kit[db] to instrument psycopg2",
        )
        psycopg2_module.Psycopg2Instrumentor().instrument()

    if enable_redis:
        redis_module = _import_optional(
            "opentelemetry.instrumentation.redis",
            "Install obs-kit[redis] to instrument Redis",
        )
        redis_module.RedisInstrumentor().instrument()

    if enable_httpx:
        httpx_module = _import_optional(
            "opentelemetry.instrumentation.httpx",
            "Install obs-kit[httpx] to instrument httpx",
        )
        httpx_module.HTTPXClientInstrumentor().instrument()


def _import_optional(module_name: str, message: str) -> Any:
    try:
        return import_module(module_name)
    except ImportError as exc:
        raise RuntimeError(message) from exc
