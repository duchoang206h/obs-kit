from __future__ import annotations

from types import SimpleNamespace
from typing import Any

import pytest

import obs_kit.auto as auto


class FakeInstrumentor:
    def __init__(self, name: str, calls: list[tuple[str, dict[str, Any]]]) -> None:
        self._name = name
        self._calls = calls

    def instrument(self, **kwargs: Any) -> None:
        self._calls.append((self._name, kwargs))


def test_instrument_python_can_enable_celery(monkeypatch: pytest.MonkeyPatch) -> None:
    calls: list[tuple[str, dict[str, Any]]] = []

    def fake_import_module(module_name: str) -> Any:
        if module_name == "opentelemetry.instrumentation.logging":
            return SimpleNamespace(
                LoggingInstrumentor=lambda: FakeInstrumentor("logging", calls),
            )
        if module_name == "opentelemetry.instrumentation.requests":
            return SimpleNamespace(
                RequestsInstrumentor=lambda: FakeInstrumentor("requests", calls),
            )
        if module_name == "opentelemetry.instrumentation.celery":
            return SimpleNamespace(
                CeleryInstrumentor=lambda: FakeInstrumentor("celery", calls),
            )
        raise ImportError(module_name)

    monkeypatch.setattr(auto, "import_module", fake_import_module)

    auto.instrument_python(enable_celery=True)

    assert calls == [
        ("logging", {"set_logging_format": False}),
        ("requests", {}),
        ("celery", {}),
    ]


def test_instrument_python_reports_missing_celery_extra(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    def fake_import_module(module_name: str) -> Any:
        if module_name == "opentelemetry.instrumentation.logging":
            return SimpleNamespace(
                LoggingInstrumentor=lambda: FakeInstrumentor("logging", []),
            )
        if module_name == "opentelemetry.instrumentation.requests":
            return SimpleNamespace(
                RequestsInstrumentor=lambda: FakeInstrumentor("requests", []),
            )
        raise ImportError(module_name)

    monkeypatch.setattr(auto, "import_module", fake_import_module)

    with pytest.raises(RuntimeError, match=r"Install obs-kit\[celery\]"):
        auto.instrument_python(enable_celery=True)
