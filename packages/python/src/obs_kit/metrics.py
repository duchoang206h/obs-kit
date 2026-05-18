from __future__ import annotations

import time
from collections.abc import Callable, Mapping
from contextlib import contextmanager
from typing import Any, TypeVar

from opentelemetry import metrics, trace
from opentelemetry.metrics import Counter, Histogram, Meter
from opentelemetry.trace import Status, StatusCode

T = TypeVar("T")

DB_CLIENT_OPERATION_DURATION = "db.client.operation.duration"
REDIS_CLIENT_OPERATION_DURATION = "redis.client.operation.duration"


def counter(
    meter: Meter,
    name: str,
    *,
    unit: str = "1",
    description: str = "",
) -> Counter:
    return meter.create_counter(name, unit=unit, description=description)


def histogram(
    meter: Meter,
    name: str,
    *,
    unit: str = "ms",
    description: str = "",
) -> Histogram:
    return meter.create_histogram(name, unit=unit, description=description)


@contextmanager
def record_db_operation(
    *,
    tracer_name: str,
    meter_name: str,
    system: str,
    operation: str,
    namespace: str | None = None,
    query_text: str | None = None,
) -> Any:
    with _record_client_operation(
        metric_name=DB_CLIENT_OPERATION_DURATION,
        metric_description="Database client operation duration.",
        tracer_name=tracer_name,
        meter_name=meter_name,
        system=system,
        operation=operation,
        namespace=namespace,
        query_text=query_text,
    ) as span:
        yield span


@contextmanager
def record_redis_operation(
    *,
    tracer_name: str,
    meter_name: str,
    command: str,
    namespace: str | None = None,
) -> Any:
    with _record_client_operation(
        metric_name=REDIS_CLIENT_OPERATION_DURATION,
        metric_description="Redis client operation duration.",
        tracer_name=tracer_name,
        meter_name=meter_name,
        system="redis",
        operation=command.upper(),
        namespace=namespace,
    ) as span:
        yield span


@contextmanager
def _record_client_operation(
    *,
    metric_name: str,
    metric_description: str,
    tracer_name: str,
    meter_name: str,
    system: str,
    operation: str,
    namespace: str | None = None,
    query_text: str | None = None,
) -> Any:
    attributes = _client_attrs(system=system, operation=operation, namespace=namespace)
    span_attributes = dict(attributes)
    if query_text:
        span_attributes["db.query.text"] = query_text

    tracer = trace.get_tracer(tracer_name)
    meter = metrics.get_meter(meter_name)
    duration = meter.create_histogram(
        metric_name,
        unit="ms",
        description=metric_description,
    )

    started = time.perf_counter()
    with tracer.start_as_current_span(
        f"{system} {operation}",
        attributes=span_attributes,
        kind=trace.SpanKind.CLIENT,
    ) as span:
        try:
            yield span
        except Exception as exc:
            span.record_exception(exc)
            span.set_status(Status(StatusCode.ERROR, str(exc)))
            duration.record(
                _elapsed_ms(started),
                attributes={**attributes, "error.type": type(exc).__name__},
            )
            raise
        duration.record(_elapsed_ms(started), attributes=attributes)


def trace_db_call(
    func: Callable[[], T],
    *,
    tracer_name: str,
    meter_name: str,
    system: str,
    operation: str,
    namespace: str | None = None,
    query_text: str | None = None,
) -> T:
    with record_db_operation(
        tracer_name=tracer_name,
        meter_name=meter_name,
        system=system,
        operation=operation,
        namespace=namespace,
        query_text=query_text,
    ):
        return func()


def trace_redis_call(
    func: Callable[[], T],
    *,
    tracer_name: str,
    meter_name: str,
    command: str,
    namespace: str | None = None,
) -> T:
    with record_redis_operation(
        tracer_name=tracer_name,
        meter_name=meter_name,
        command=command,
        namespace=namespace,
    ):
        return func()


def _client_attrs(
    *,
    system: str,
    operation: str,
    namespace: str | None,
) -> Mapping[str, str]:
    attributes = {
        "db.system": system,
        "db.operation.name": operation,
    }
    if namespace:
        attributes["db.namespace"] = namespace
    return attributes


def _elapsed_ms(started: float) -> float:
    return (time.perf_counter() - started) * 1000
