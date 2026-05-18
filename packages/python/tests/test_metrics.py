from __future__ import annotations

import pytest

import obs_kit.metrics as kit_metrics


class FakeHistogram:
    def __init__(self) -> None:
        self.records: list[tuple[float, dict[str, str]]] = []

    def record(self, value: float, *, attributes: dict[str, str]) -> None:
        self.records.append((value, attributes))


class FakeMeter:
    def __init__(self, histogram: FakeHistogram) -> None:
        self.histogram = histogram
        self.histogram_names: list[str] = []

    def create_histogram(self, name: str, **_kwargs: object) -> FakeHistogram:
        self.histogram_names.append(name)
        return self.histogram


class FakeSpan:
    def __init__(self) -> None:
        self.exceptions: list[Exception] = []
        self.status: object | None = None

    def __enter__(self) -> FakeSpan:
        return self

    def __exit__(self, *_args: object) -> None:
        return None

    def record_exception(self, exc: Exception) -> None:
        self.exceptions.append(exc)

    def set_status(self, status: object) -> None:
        self.status = status


class FakeTracer:
    def __init__(self, span: FakeSpan) -> None:
        self.span = span
        self.names: list[str] = []
        self.attributes: list[dict[str, str]] = []

    def start_as_current_span(self, name: str, **kwargs: object) -> FakeSpan:
        self.names.append(name)
        self.attributes.append(kwargs["attributes"])  # type: ignore[arg-type]
        return self.span


def test_record_db_operation_records_duration_and_attrs(monkeypatch: pytest.MonkeyPatch) -> None:
    histogram = FakeHistogram()
    span = FakeSpan()
    tracer = FakeTracer(span)
    meter = FakeMeter(histogram)
    monkeypatch.setattr("obs_kit.metrics.metrics.get_meter", lambda _name: meter)
    monkeypatch.setattr("obs_kit.metrics.trace.get_tracer", lambda _name: tracer)

    with kit_metrics.record_db_operation(
        tracer_name="orders",
        meter_name="orders",
        system="postgresql",
        operation="SELECT",
        namespace="orders",
        query_text="SELECT * FROM orders",
    ):
        pass

    assert tracer.names == ["postgresql SELECT"]
    assert meter.histogram_names == [kit_metrics.DB_CLIENT_OPERATION_DURATION]
    assert tracer.attributes[0]["db.system"] == "postgresql"
    assert tracer.attributes[0]["db.operation.name"] == "SELECT"
    assert tracer.attributes[0]["db.namespace"] == "orders"
    assert tracer.attributes[0]["db.query.text"] == "SELECT * FROM orders"
    assert histogram.records[0][1] == {
        "db.system": "postgresql",
        "db.operation.name": "SELECT",
        "db.namespace": "orders",
    }


def test_record_redis_operation_records_error_type(monkeypatch: pytest.MonkeyPatch) -> None:
    histogram = FakeHistogram()
    span = FakeSpan()
    meter = FakeMeter(histogram)
    monkeypatch.setattr("obs_kit.metrics.metrics.get_meter", lambda _name: meter)
    monkeypatch.setattr("obs_kit.metrics.trace.get_tracer", lambda _name: FakeTracer(span))

    with pytest.raises(ValueError, match="failed"):
        with kit_metrics.record_redis_operation(
            tracer_name="orders",
            meter_name="orders",
            command="get",
        ):
            raise ValueError("failed")

    assert span.exceptions
    assert meter.histogram_names == [kit_metrics.REDIS_CLIENT_OPERATION_DURATION]
    assert histogram.records[0][1]["db.system"] == "redis"
    assert histogram.records[0][1]["db.operation.name"] == "GET"
    assert histogram.records[0][1]["error.type"] == "ValueError"
