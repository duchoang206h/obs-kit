package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/duchoang206h/obs-kit/packages/go/observability"
)

func main() {
	ctx := context.Background()
	obs, err := observability.Init(ctx, &observability.Config{
		ServiceName:    "go-http",
		ServiceVersion: "0.1.0",
		Environment:    "local",
		Logging: observability.LoggingConfig{
			Level: "info",
		},
	})
	if err != nil {
		log.Fatalf("init observability: %v", err)
	}
	defer obs.Shutdown(ctx)

	logger, file, err := observability.NewFileLogger(&obs.Config, "../../logs/go-http.log")
	if err != nil {
		log.Fatalf("create logger: %v", err)
	}
	defer file.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/orders", func(writer http.ResponseWriter, request *http.Request) {
		observability.InfoContext(
			request.Context(),
			logger,
			"order created",
			slog.String("order.id", "ord_http_go"),
		)
		writer.WriteHeader(http.StatusCreated)
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           observability.HTTPMiddleware(&obs.Config, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("listening on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}
