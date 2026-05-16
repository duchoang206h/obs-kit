import json
import logging
from pathlib import Path
from typing import Any

from obs_kit.config import Config
from obs_kit.logging import JsonFormatter


def test_json_formatter_adds_service_fields() -> None:
    config = Config(service_name="orders", service_version="1.2.3", environment="test")
    record = logging.LogRecord(
        name="orders.worker",
        level=logging.INFO,
        pathname=__file__,
        lineno=10,
        msg="order created",
        args=(),
        exc_info=None,
    )
    record.trace_id = "abc"
    record.span_id = "def"

    payload = json.loads(JsonFormatter(config).format(record))

    assert payload["level"] == "info"
    assert payload["message"] == "order created"
    assert payload["service.name"] == "orders"
    assert payload["service.version"] == "1.2.3"
    assert payload["deployment.environment"] == "test"
    assert payload["trace_id"] == "abc"
    assert payload["span_id"] == "def"


def test_json_formatter_matches_log_contract_fixture() -> None:
    fixture = _fixture("logs", "order-created.json")
    config = Config(
        service_name=fixture["service.name"],
        service_version=fixture["service.version"],
        environment=fixture["deployment.environment"],
    )
    record = logging.LogRecord(
        name="orders.worker",
        level=logging.INFO,
        pathname=__file__,
        lineno=10,
        msg=fixture["message"],
        args=(),
        exc_info=None,
    )
    record.trace_id = fixture["trace_id"]
    record.span_id = fixture["span_id"]
    record.__dict__["order.id"] = fixture["order.id"]

    payload = json.loads(JsonFormatter(config).format(record))

    for key, value in fixture.items():
        assert payload[key] == value


def test_json_formatter_adds_error_fields() -> None:
    config = Config(service_name="orders", service_version="1.2.3", environment="test")

    try:
        raise ValueError("bad order")
    except ValueError as error:
        record = logging.LogRecord(
            name="orders.worker",
            level=logging.ERROR,
            pathname=__file__,
            lineno=10,
            msg="failed",
            args=(),
            exc_info=(type(error), error, error.__traceback__),
        )

    payload = json.loads(JsonFormatter(config).format(record))

    assert payload["error.type"] == "ValueError"
    assert payload["error.message"] == "bad order"
    assert "ValueError: bad order" in payload["error.stack"]


def _fixture(*path: str) -> dict[str, Any]:
    fixture_path = Path(__file__).resolve().parents[3] / "contracts" / "fixtures" / Path(*path)
    data: dict[str, Any] = json.loads(fixture_path.read_text())
    return data
