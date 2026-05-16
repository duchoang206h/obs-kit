from __future__ import annotations

import json
import logging
import traceback
from collections.abc import Mapping
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

from opentelemetry import trace

from obs_kit.config import Config, resolve_config

_STANDARD_FIELDS = {
    "args",
    "asctime",
    "created",
    "exc_info",
    "exc_text",
    "filename",
    "funcName",
    "levelname",
    "levelno",
    "lineno",
    "module",
    "msecs",
    "message",
    "msg",
    "name",
    "pathname",
    "process",
    "processName",
    "relativeCreated",
    "stack_info",
    "taskName",
    "thread",
    "threadName",
}


class JsonFormatter(logging.Formatter):
    def __init__(self, config: Config) -> None:
        super().__init__()
        self._config = config

    def format(self, record: logging.LogRecord) -> str:
        payload: dict[str, Any] = {
            "timestamp": datetime.fromtimestamp(record.created, UTC).isoformat(),
            "level": record.levelname.lower(),
            "message": record.getMessage(),
            "logger": record.name,
            "service.name": self._config.service_name or "unknown-service",
            "service.version": self._config.service_version or "unknown",
            "deployment.environment": self._config.environment or "development",
        }
        span_context = trace.get_current_span().get_span_context()
        if span_context.is_valid:
            payload["trace_id"] = trace.format_trace_id(span_context.trace_id)
            payload["span_id"] = trace.format_span_id(span_context.span_id)

        for key, value in record.__dict__.items():
            if key not in _STANDARD_FIELDS:
                payload[key] = value

        if record.exc_info:
            error_type = record.exc_info[0]
            error_value = record.exc_info[1]
            payload["error.type"] = error_type.__name__ if error_type else "Error"
            payload["error.message"] = str(error_value) if error_value else ""
            payload["error.stack"] = self.formatException(record.exc_info)
        elif "error" in payload and isinstance(payload["error"], BaseException):
            error = payload.pop("error")
            payload["error.type"] = type(error).__name__
            payload["error.message"] = str(error)
            payload["error.stack"] = "".join(
                traceback.format_exception(type(error), error, error.__traceback__)
            )

        return json.dumps(payload, default=str, separators=(",", ":"))


def configure_logging(
    config: Config | None = None,
    *,
    log_file: str | Path | None = None,
    logger_name: str = "",
) -> logging.Logger:
    resolved = resolve_config(config)
    logger = logging.getLogger(logger_name)
    logger.setLevel(_level(resolved.logging.level or "info"))
    logger.handlers.clear()
    logger.propagate = False

    handler: logging.Handler
    if log_file is None:
        handler = logging.StreamHandler()
    else:
        path = Path(log_file)
        path.parent.mkdir(parents=True, exist_ok=True)
        handler = logging.FileHandler(path)

    handler.setFormatter(JsonFormatter(resolved))
    logger.addHandler(handler)
    return logger


def get_logger(name: str, config: Config | None = None) -> logging.Logger:
    logger = logging.getLogger(name)
    if not logger.handlers:
        configure_logging(config, logger_name=name)
    return logger


def log_extra(**fields: Any) -> Mapping[str, Any]:
    return {"extra": fields}


def _level(level: str) -> int:
    return getattr(logging, level.upper(), logging.INFO)
