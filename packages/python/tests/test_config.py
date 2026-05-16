import json
from pathlib import Path
from typing import Any

from obs_kit import Config, TracingConfig, resolve_config


def test_resolve_config_uses_environment(monkeypatch: Any) -> None:
    fixture = _fixture("config", "env-basic.json")
    for key, value in fixture["env"].items():
        monkeypatch.setenv(key, value)

    config = resolve_config()
    expected = fixture["expected"]

    assert config.service_name == expected["serviceName"]
    assert config.service_version == expected["serviceVersion"]
    assert config.environment == expected["environment"]
    assert config.collector.url == expected["collectorUrl"]
    assert config.tracing.sample_rate == expected["tracingSampleRate"]
    assert config.metrics.export_interval == expected["metricsExportInterval"]
    assert config.logging.level == expected["loggingLevel"]


def test_resolve_config_merges_explicit_values_with_environment(monkeypatch: Any) -> None:
    monkeypatch.setenv("OTEL_SERVICE_NAME", "env-service")
    monkeypatch.setenv("ENVIRONMENT", "test")
    monkeypatch.setenv("LOG_LEVEL", "debug")

    config = resolve_config(
        Config(service_name="explicit-service", tracing=TracingConfig(enabled=False))
    )

    assert config.service_name == "explicit-service"
    assert config.environment == "test"
    assert config.logging.level == "debug"
    assert config.tracing.enabled is False


def test_resolve_config_preserves_explicit_zero_sample_rate(monkeypatch: Any) -> None:
    monkeypatch.setenv("OTEL_TRACES_SAMPLER_ARG", "1.0")

    config = resolve_config(Config(tracing=TracingConfig(sample_rate=0.0)))

    assert config.tracing.sample_rate == 0.0


def _fixture(*path: str) -> dict[str, Any]:
    fixture_path = Path(__file__).resolve().parents[3] / "contracts" / "fixtures" / Path(*path)
    data: dict[str, Any] = json.loads(fixture_path.read_text())
    return data
