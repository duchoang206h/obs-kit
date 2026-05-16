from __future__ import annotations

from dataclasses import dataclass

from opentelemetry import metrics, trace
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.sdk.resources import (
    DEPLOYMENT_ENVIRONMENT,
    SERVICE_NAME,
    SERVICE_VERSION,
    Resource,
)
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.trace.sampling import ParentBased, TraceIdRatioBased

from obs_kit.config import Config, resolve_config


@dataclass(frozen=True)
class Observability:
    config: Config
    tracer_provider: TracerProvider | None
    meter_provider: MeterProvider | None

    def tracer(self, name: str) -> trace.Tracer:
        return trace.get_tracer(name)

    def meter(self, name: str) -> metrics.Meter:
        return metrics.get_meter(name)

    def shutdown(self) -> None:
        if self.tracer_provider is not None:
            self.tracer_provider.shutdown()
        if self.meter_provider is not None:
            self.meter_provider.shutdown()


def init_observability(config: Config | None = None) -> Observability:
    resolved = resolve_config(config)
    resource = Resource.create(
        {
            SERVICE_NAME: resolved.service_name or "unknown-service",
            SERVICE_VERSION: resolved.service_version or "unknown",
            DEPLOYMENT_ENVIRONMENT: resolved.environment or "development",
        }
    )

    tracer_provider: TracerProvider | None = None
    if resolved.tracing.enabled:
        sample_rate = (
            1.0 if resolved.tracing.sample_rate is None else resolved.tracing.sample_rate
        )
        tracer_provider = TracerProvider(
            resource=resource,
            sampler=ParentBased(TraceIdRatioBased(sample_rate)),
        )
        tracer_provider.add_span_processor(
            BatchSpanProcessor(OTLPSpanExporter(endpoint=_otlp_endpoint(resolved, "v1/traces")))
        )
        trace.set_tracer_provider(tracer_provider)

    meter_provider: MeterProvider | None = None
    if resolved.metrics.enabled:
        metric_reader = PeriodicExportingMetricReader(
            OTLPMetricExporter(endpoint=_otlp_endpoint(resolved, "v1/metrics")),
            export_interval_millis=resolved.metrics.export_interval or 60000,
        )
        meter_provider = MeterProvider(resource=resource, metric_readers=[metric_reader])
        metrics.set_meter_provider(meter_provider)

    return Observability(
        config=resolved,
        tracer_provider=tracer_provider,
        meter_provider=meter_provider,
    )


def shutdown_observability(observability: Observability) -> None:
    observability.shutdown()


def _otlp_endpoint(config: Config, path: str) -> str:
    base = (config.collector.url or "http://localhost:4318").rstrip("/")
    return f"{base}/{path}"
