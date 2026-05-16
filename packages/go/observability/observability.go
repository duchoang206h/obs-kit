package observability

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

type Observability struct {
	Config         Config
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
}

func Init(ctx context.Context, config *Config) (*Observability, error) {
	resolved := ResolveConfig(config)
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(resolved.ServiceName),
		semconv.ServiceVersion(resolved.ServiceVersion),
		semconv.DeploymentEnvironmentName(resolved.Environment),
	)

	obs := &Observability{Config: resolved}

	if boolValue(resolved.Tracing.Enabled) {
		traceExporter, err := otlptracehttp.New(ctx, otlpHTTPOptions(resolved)...)
		if err != nil {
			return nil, err
		}

		obs.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(floatValue(resolved.Tracing.SampleRate, 1.0)))),
			sdktrace.WithBatcher(traceExporter),
		)
		otel.SetTracerProvider(obs.tracerProvider)
	}

	if boolValue(resolved.Metrics.Enabled) {
		metricExporter, err := otlpmetrichttp.New(ctx, otlpMetricHTTPOptions(resolved)...)
		if err != nil {
			return nil, err
		}

		obs.meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
				metricExporter,
				sdkmetric.WithInterval(time.Duration(resolved.Metrics.ExportInterval)*time.Millisecond),
			)),
		)
		otel.SetMeterProvider(obs.meterProvider)
	}

	return obs, nil
}

func (o Observability) Logger() *slog.Logger {
	return NewLogger(&o.Config, nil)
}

func (o Observability) Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func (o Observability) Meter(name string) metric.Meter {
	return otel.Meter(name)
}

func (o Observability) Shutdown(ctx context.Context) error {
	if o.tracerProvider != nil {
		if err := o.tracerProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if o.meterProvider != nil {
		if err := o.meterProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

func otlpHTTPOptions(config Config) []otlptracehttp.Option {
	endpoint := strings.TrimPrefix(config.Collector.URL, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	options := []otlptracehttp.Option{otlptracehttp.WithEndpoint(endpoint)}
	if strings.HasPrefix(config.Collector.URL, "http://") {
		options = append(options, otlptracehttp.WithInsecure())
	}
	return options
}

func otlpMetricHTTPOptions(config Config) []otlpmetrichttp.Option {
	endpoint := strings.TrimPrefix(config.Collector.URL, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	options := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(endpoint)}
	if strings.HasPrefix(config.Collector.URL, "http://") {
		options = append(options, otlpmetrichttp.WithInsecure())
	}
	return options
}
