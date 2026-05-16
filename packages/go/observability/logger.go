package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

func NewLogger(config *Config, writer io.Writer) *slog.Logger {
	resolved := ResolveConfig(config)
	if writer == nil {
		writer = os.Stdout
	}

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: parseLevel(resolved.Logging.Level),
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			switch attr.Key {
			case slog.TimeKey:
				attr.Key = "timestamp"
			case slog.LevelKey:
				attr.Value = slog.StringValue(strings.ToLower(attr.Value.String()))
			case slog.MessageKey:
				attr.Key = "message"
			}
			return attr
		},
	})

	return slog.New(handler).With(
		"service.name", resolved.ServiceName,
		"service.version", resolved.ServiceVersion,
		"deployment.environment", resolved.Environment,
	)
}

func NewFileLogger(config *Config, path string) (*slog.Logger, *os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, err
	}

	return NewLogger(config, file), file, nil
}

func InfoContext(ctx context.Context, logger *slog.Logger, message string, args ...any) {
	logger.InfoContext(ctx, message, appendTraceAttrs(ctx, args)...)
}

func ErrorContext(ctx context.Context, logger *slog.Logger, message string, args ...any) {
	logger.ErrorContext(ctx, message, appendTraceAttrs(ctx, args)...)
}

func appendTraceAttrs(ctx context.Context, args []any) []any {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return args
	}

	return append(
		args,
		slog.String("trace_id", spanContext.TraceID().String()),
		slog.String("span_id", spanContext.SpanID().String()),
	)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
