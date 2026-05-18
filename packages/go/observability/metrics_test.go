package observability

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

func TestClientOperationAttrs(t *testing.T) {
	attrs := clientOperationAttrs(ClientOperation{
		System:    "postgresql",
		Operation: "SELECT",
		Namespace: "orders",
	})

	got := map[string]string{}
	for _, attr := range attrs {
		got[string(attr.Key)] = attr.Value.AsString()
	}

	if got["db.system"] != "postgresql" {
		t.Fatalf("expected db.system attr, got %q", got["db.system"])
	}
	if got["db.operation.name"] != "SELECT" {
		t.Fatalf("expected db.operation.name attr, got %q", got["db.operation.name"])
	}
	if got["db.namespace"] != "orders" {
		t.Fatalf("expected db.namespace attr, got %q", got["db.namespace"])
	}
}

func TestTraceRedisOperationReturnsCallbackError(t *testing.T) {
	tracer := tracenoop.NewTracerProvider().Tracer("test")
	meter := noop.NewMeterProvider().Meter("test")
	expected := errors.New("redis failed")

	err := TraceRedisOperation(
		context.Background(),
		tracer,
		meter,
		"GET",
		"0",
		func(context.Context) error {
			return expected
		},
	)

	if !errors.Is(err, expected) {
		t.Fatalf("expected callback error, got %v", err)
	}
}

func TestTraceDBOperationRePanics(t *testing.T) {
	tracer := tracenoop.NewTracerProvider().Tracer("test")
	meter := noop.NewMeterProvider().Meter("test")

	defer func() {
		recovered := recover()
		if recovered != "boom" {
			t.Fatalf("expected panic to be rethrown, got %v", recovered)
		}
	}()

	_ = TraceDBOperation(
		context.Background(),
		tracer,
		meter,
		ClientOperation{System: "postgresql", Operation: "SELECT"},
		func(context.Context) error {
			panic("boom")
		},
	)
}

func TestErrorType(t *testing.T) {
	err := &typedTestError{}

	if got := errorType(err); got != "github.com/duchoang206h/obs-kit/packages/go/observability.typedTestError" {
		t.Fatalf("unexpected error type: %s", got)
	}
}

type typedTestError struct{}

func (e *typedTestError) Error() string {
	return "typed"
}
