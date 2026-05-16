package observability

import (
	"bufio"
	"net"
	"net/http"
	"slices"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

func HTTPMiddleware(config *Config, next http.Handler) http.Handler {
	resolved := ResolveConfig(config)
	tracer := otel.Tracer(resolved.ServiceName)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if slices.Contains(resolved.Tracing.IgnorePaths, request.URL.Path) {
			next.ServeHTTP(writer, request)
			return
		}

		started := time.Now()
		spanName := request.Method + " " + request.URL.Path
		ctx := otel.GetTextMapPropagator().Extract(
			request.Context(),
			propagation.HeaderCarrier(request.Header),
		)
		if !trace.SpanContextFromContext(ctx).IsValid() {
			ctx = propagation.TraceContext{}.Extract(
				request.Context(),
				propagation.HeaderCarrier(request.Header),
			)
		}
		ctx, span := tracer.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(request.Method),
				semconv.URLPath(request.URL.Path),
			),
		)
		defer span.End()

		recorder, wrappedWriter := newStatusRecorder(writer)
		next.ServeHTTP(wrappedWriter, request.WithContext(ctx))

		span.SetAttributes(
			semconv.HTTPResponseStatusCode(recorder.statusCode),
			attribute.Int64("http.server.duration_ms", time.Since(started).Milliseconds()),
		)
		if recorder.statusCode >= http.StatusInternalServerError {
			span.RecordError(httpError{statusCode: recorder.statusCode})
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newStatusRecorder(writer http.ResponseWriter) (*statusRecorder, http.ResponseWriter) {
	recorder := &statusRecorder{ResponseWriter: writer, statusCode: http.StatusOK}
	_, flusher := writer.(http.Flusher)
	_, hijacker := writer.(http.Hijacker)
	_, pusher := writer.(http.Pusher)

	switch {
	case flusher && hijacker && pusher:
		return recorder, &flushHijackPushRecorder{statusRecorder: recorder}
	case flusher && hijacker:
		return recorder, &flushHijackRecorder{statusRecorder: recorder}
	case flusher && pusher:
		return recorder, &flushPushRecorder{statusRecorder: recorder}
	case hijacker && pusher:
		return recorder, &hijackPushRecorder{statusRecorder: recorder}
	case flusher:
		return recorder, &flushRecorder{statusRecorder: recorder}
	case hijacker:
		return recorder, &hijackRecorder{statusRecorder: recorder}
	case pusher:
		return recorder, &pushRecorder{statusRecorder: recorder}
	default:
		return recorder, recorder
	}
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	return r.ResponseWriter.Write(data)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

type flushRecorder struct {
	*statusRecorder
}

func (r *flushRecorder) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

type hijackRecorder struct {
	*statusRecorder
}

func (r *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

type pushRecorder struct {
	*statusRecorder
}

func (r *pushRecorder) Push(target string, opts *http.PushOptions) error {
	return r.ResponseWriter.(http.Pusher).Push(target, opts)
}

type flushHijackRecorder struct {
	*statusRecorder
}

func (r *flushHijackRecorder) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

func (r *flushHijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

type flushPushRecorder struct {
	*statusRecorder
}

func (r *flushPushRecorder) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

func (r *flushPushRecorder) Push(target string, opts *http.PushOptions) error {
	return r.ResponseWriter.(http.Pusher).Push(target, opts)
}

type hijackPushRecorder struct {
	*statusRecorder
}

func (r *hijackPushRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

func (r *hijackPushRecorder) Push(target string, opts *http.PushOptions) error {
	return r.ResponseWriter.(http.Pusher).Push(target, opts)
}

type flushHijackPushRecorder struct {
	*statusRecorder
}

func (r *flushHijackPushRecorder) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

func (r *flushHijackPushRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

func (r *flushHijackPushRecorder) Push(target string, opts *http.PushOptions) error {
	return r.ResponseWriter.(http.Pusher).Push(target, opts)
}

type httpError struct {
	statusCode int
}

func (e httpError) Error() string {
	return "http status " + strconv.Itoa(e.statusCode)
}
