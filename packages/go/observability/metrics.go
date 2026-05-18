package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	DBClientOperationDuration    = "db.client.operation.duration"
	RedisClientOperationDuration = "redis.client.operation.duration"
)

type ClientOperation struct {
	System    string
	Operation string
	Namespace string
	QueryText string
}

func NewCounter(meter metric.Meter, name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return meter.Int64Counter(name, opts...)
}

func NewDurationHistogram(meter metric.Meter, name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return meter.Float64Histogram(
		name,
		append(
			[]metric.Float64HistogramOption{
				metric.WithUnit("ms"),
			},
			opts...,
		)...,
	)
}

func TraceDBOperation(
	ctx context.Context,
	tracer trace.Tracer,
	meter metric.Meter,
	operation ClientOperation,
	fn func(context.Context) error,
) error {
	histogram, err := NewDurationHistogram(
		meter,
		DBClientOperationDuration,
		metric.WithDescription("Database client operation duration."),
	)
	if err != nil {
		return err
	}

	return traceClientOperation(ctx, tracer, histogram, operation, fn)
}

func TraceRedisOperation(
	ctx context.Context,
	tracer trace.Tracer,
	meter metric.Meter,
	command string,
	namespace string,
	fn func(context.Context) error,
) error {
	histogram, err := NewDurationHistogram(
		meter,
		RedisClientOperationDuration,
		metric.WithDescription("Redis client operation duration."),
	)
	if err != nil {
		return err
	}

	return traceClientOperation(ctx, tracer, histogram, ClientOperation{
		System:    "redis",
		Operation: command,
		Namespace: namespace,
	}, fn)
}

func traceClientOperation(
	ctx context.Context,
	tracer trace.Tracer,
	histogram metric.Float64Histogram,
	operation ClientOperation,
	fn func(context.Context) error,
) (err error) {
	attrs := clientOperationAttrs(operation)
	spanAttrs := attrs
	if operation.QueryText != "" {
		spanAttrs = append(spanAttrs, attribute.String("db.query.text", operation.QueryText))
	}

	ctx, span := tracer.Start(
		ctx,
		operation.System+" "+operation.Operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(spanAttrs...),
	)
	started := time.Now()
	defer func() {
		if recovered := recover(); recovered != nil {
			panicErr := fmt.Errorf("panic: %v", recovered)
			span.RecordError(panicErr)
			span.SetStatus(codes.Error, panicErr.Error())
			attrs = append(attrs, attribute.String("error.type", errorType(panicErr)))
			histogram.Record(ctx, elapsedMilliseconds(started), metric.WithAttributes(attrs...))
			span.End()
			panic(recovered)
		}

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			attrs = append(attrs, attribute.String("error.type", errorType(err)))
		}
		histogram.Record(ctx, elapsedMilliseconds(started), metric.WithAttributes(attrs...))
		span.End()
	}()

	err = fn(ctx)
	return err
}

func clientOperationAttrs(operation ClientOperation) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("db.system", operation.System),
		attribute.String("db.operation.name", operation.Operation),
	}
	if operation.Namespace != "" {
		attrs = append(attrs, attribute.String("db.namespace", operation.Namespace))
	}
	return attrs
}

func elapsedMilliseconds(started time.Time) float64 {
	return float64(time.Since(started).Microseconds()) / 1000
}
