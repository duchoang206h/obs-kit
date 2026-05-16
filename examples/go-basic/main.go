package main

import (
	"log"
	"log/slog"
	"path/filepath"

	"github.com/duchoang206h/obs-kit/packages/go/observability"
)

func main() {
	logPath := filepath.Join("..", "..", "logs", "go-basic.log")
	logger, file, err := observability.NewFileLogger(&observability.Config{
		ServiceName:    "go-basic",
		ServiceVersion: "0.1.0",
		Environment:    "local",
		Logging: observability.LoggingConfig{
			Level: "info",
		},
	}, logPath)
	if err != nil {
		log.Fatalf("create logger: %v", err)
	}
	defer file.Close()

	logger.Info(
		"order created",
		slog.String("trace_id", "demo-go-trace"),
		slog.String("span_id", "demo-go-span"),
		slog.String("order.id", "ord_789"),
		slog.String("customer.id", "cus_012"),
	)
}
