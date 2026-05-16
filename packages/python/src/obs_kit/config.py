from __future__ import annotations

import os
from dataclasses import dataclass, field
from typing import TypeVar

T = TypeVar("T")


@dataclass(frozen=True)
class CollectorConfig:
    url: str | None = None


@dataclass(frozen=True)
class TracingConfig:
    enabled: bool | None = None
    sample_rate: float | None = None
    ignore_paths: tuple[str, ...] = ("/health", "/metrics")


@dataclass(frozen=True)
class MetricsConfig:
    enabled: bool | None = None
    export_interval: int | None = None


@dataclass(frozen=True)
class LoggingConfig:
    level: str | None = None


@dataclass(frozen=True)
class Config:
    service_name: str | None = None
    service_version: str | None = None
    environment: str | None = None
    collector: CollectorConfig = field(default_factory=CollectorConfig)
    tracing: TracingConfig = field(default_factory=TracingConfig)
    metrics: MetricsConfig = field(default_factory=MetricsConfig)
    logging: LoggingConfig = field(default_factory=LoggingConfig)


def resolve_config(config: Config | None = None) -> Config:
    service_name = _get_env("OTEL_SERVICE_NAME", default="unknown-service")
    environment = _get_env("ENVIRONMENT", default="development")

    env_config = Config(
        service_name=service_name,
        service_version=_get_env("OTEL_SERVICE_VERSION", default="unknown"),
        environment=environment,
        collector=CollectorConfig(
            url=_get_env("OTEL_EXPORTER_OTLP_ENDPOINT", default="http://localhost:4318"),
        ),
        tracing=TracingConfig(
            enabled=True,
            sample_rate=_get_float("OTEL_TRACES_SAMPLER_ARG", 1.0),
            ignore_paths=("/health", "/metrics"),
        ),
        metrics=MetricsConfig(
            enabled=True,
            export_interval=_get_int("OTEL_METRICS_EXPORT_INTERVAL", 60000),
        ),
        logging=LoggingConfig(level=_get_env("LOG_LEVEL", default="info")),
    )

    if config is None:
        return env_config

    return Config(
        service_name=config.service_name or env_config.service_name,
        service_version=config.service_version or env_config.service_version,
        environment=config.environment or env_config.environment,
        collector=CollectorConfig(
            url=config.collector.url or env_config.collector.url,
        ),
        tracing=TracingConfig(
            enabled=_coalesce(config.tracing.enabled, env_config.tracing.enabled),
            sample_rate=_coalesce(config.tracing.sample_rate, env_config.tracing.sample_rate),
            ignore_paths=config.tracing.ignore_paths or env_config.tracing.ignore_paths,
        ),
        metrics=MetricsConfig(
            enabled=_coalesce(config.metrics.enabled, env_config.metrics.enabled),
            export_interval=_coalesce(
                config.metrics.export_interval,
                env_config.metrics.export_interval,
            ),
        ),
        logging=LoggingConfig(
            level=config.logging.level or env_config.logging.level,
        ),
    )


def _coalesce(value: T | None, fallback: T) -> T:
    return fallback if value is None else value


def _get_env(*keys: str, default: str) -> str:
    for key in keys:
        value = os.getenv(key)
        if value:
            return value
    return default


def _get_float(key: str, default: float) -> float:
    value = os.getenv(key)
    if not value:
        return default
    try:
        return float(value)
    except ValueError:
        return default


def _get_int(key: str, default: int) -> int:
    value = os.getenv(key)
    if not value:
        return default
    try:
        return int(value)
    except ValueError:
        return default
