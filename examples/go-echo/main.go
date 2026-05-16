package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/duchoang206h/obs-kit/packages/go/observability"
	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.Background()
	obs, err := observability.Init(ctx, &observability.Config{
		ServiceName:    "go-echo",
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

	logger, file, err := observability.NewFileLogger(&obs.Config, "../../logs/go-echo.log")
	if err != nil {
		log.Fatalf("create logger: %v", err)
	}
	defer file.Close()

	app := echo.New()
	app.HideBanner = true
	app.HidePort = true
	app.POST("/orders", func(c echo.Context) error {
		observability.InfoContext(
			c.Request().Context(),
			logger,
			"order created",
			slog.String("order.id", "ord_echo_go"),
		)
		return c.JSON(http.StatusCreated, map[string]string{"order_id": "ord_echo_go"})
	})

	server := &http.Server{
		Addr:              ":8082",
		Handler:           observability.HTTPMiddleware(&obs.Config, app),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("listening on http://localhost:8082")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}
