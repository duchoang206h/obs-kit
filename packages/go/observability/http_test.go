package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestHTTPMiddlewareSkipsIgnoredPaths(t *testing.T) {
	called := false
	handler := HTTPMiddleware(&Config{
		ServiceName: "orders",
		Tracing: TracingConfig{
			IgnorePaths: []string{"/health"},
		},
	}, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		called = true
		writer.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}
}

func TestHTTPMiddlewareExtractsIncomingTraceContext(t *testing.T) {
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})

	provider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
	)
	defer provider.Shutdown(t.Context())
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

	parentTraceID := "0af7651916cd43dd8448eb211c80319c"
	var gotTraceID string
	var gotRemoteParent bool

	handler := HTTPMiddleware(&Config{
		ServiceName: "orders",
	}, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		spanContext := oteltrace.SpanContextFromContext(request.Context())
		gotTraceID = spanContext.TraceID().String()
		gotRemoteParent = spanContext.IsRemote()
		writer.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/orders", nil)
	request.Header.Set("traceparent", "00-"+parentTraceID+"-b7ad6b7169203331-01")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if gotTraceID != parentTraceID {
		t.Fatalf("expected trace ID %s, got %s", parentTraceID, gotTraceID)
	}
	if gotRemoteParent {
		t.Fatal("expected handler context to contain the server span, not the remote parent")
	}
}

func TestStatusRecorderSupportsResponseControllerFlush(t *testing.T) {
	flushed := false
	recorder := &statusRecorder{
		ResponseWriter: flushResponseWriter{
			ResponseWriter: httptest.NewRecorder(),
			flush: func() {
				flushed = true
			},
		},
	}

	controller := http.NewResponseController(recorder)
	if err := controller.Flush(); err != nil {
		t.Fatalf("expected flush support through wrapped writer: %v", err)
	}
	if !flushed {
		t.Fatal("expected wrapped flusher to be called")
	}
}

func TestStatusRecorderOnlyExposesSupportedOptionalInterfaces(t *testing.T) {
	_, plainWriter := newStatusRecorder(plainResponseWriter{header: http.Header{}})
	if _, ok := plainWriter.(http.Flusher); ok {
		t.Fatal("expected plain writer wrapper not to expose http.Flusher")
	}
	if _, ok := plainWriter.(http.Hijacker); ok {
		t.Fatal("expected plain writer wrapper not to expose http.Hijacker")
	}

	flushed := false
	_, flushWriter := newStatusRecorder(flushResponseWriter{
		ResponseWriter: httptest.NewRecorder(),
		flush: func() {
			flushed = true
		},
	})

	flusher, ok := flushWriter.(http.Flusher)
	if !ok {
		t.Fatal("expected wrapper to expose http.Flusher")
	}
	flusher.Flush()
	if !flushed {
		t.Fatal("expected wrapped flusher to be called")
	}
	if _, ok := flushWriter.(http.Hijacker); ok {
		t.Fatal("expected flush-only wrapper not to expose http.Hijacker")
	}
}

type flushResponseWriter struct {
	http.ResponseWriter
	flush func()
}

func (w flushResponseWriter) Flush() {
	w.flush()
}

type plainResponseWriter struct {
	header http.Header
}

func (w plainResponseWriter) Header() http.Header {
	return w.header
}

func (w plainResponseWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (w plainResponseWriter) WriteHeader(statusCode int) {}
