package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoggerAddsServiceFields(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&Config{
		ServiceName:    "orders",
		ServiceVersion: "1.2.3",
		Environment:    "test",
		Logging: LoggingConfig{
			Level: "info",
		},
	}, &buffer)

	logger.Info("order created", slog.String("trace_id", "abc"), slog.String("span_id", "def"))

	var payload map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON log payload: %v", err)
	}

	if payload["message"] != "order created" {
		t.Fatalf("expected message order created, got %v", payload["message"])
	}
	if payload["level"] != "info" {
		t.Fatalf("expected level info, got %v", payload["level"])
	}
	if payload["service.name"] != "orders" {
		t.Fatalf("expected service name orders, got %v", payload["service.name"])
	}
	if payload["service.version"] != "1.2.3" {
		t.Fatalf("expected service version 1.2.3, got %v", payload["service.version"])
	}
	if payload["deployment.environment"] != "test" {
		t.Fatalf("expected environment test, got %v", payload["deployment.environment"])
	}
	if payload["trace_id"] != "abc" {
		t.Fatalf("expected trace_id abc, got %v", payload["trace_id"])
	}
	if payload["span_id"] != "def" {
		t.Fatalf("expected span_id def, got %v", payload["span_id"])
	}
}

func TestNewLoggerMatchesLogContractFixture(t *testing.T) {
	fixture := loadLogFixture(t)
	var buffer bytes.Buffer
	logger := NewLogger(&Config{
		ServiceName:    fixture["service.name"].(string),
		ServiceVersion: fixture["service.version"].(string),
		Environment:    fixture["deployment.environment"].(string),
		Logging: LoggingConfig{
			Level: "info",
		},
	}, &buffer)

	logger.Info(
		fixture["message"].(string),
		slog.String("trace_id", fixture["trace_id"].(string)),
		slog.String("span_id", fixture["span_id"].(string)),
		slog.String("order.id", fixture["order.id"].(string)),
	)

	var payload map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON log payload: %v", err)
	}

	for key, value := range fixture {
		if payload[key] != value {
			t.Fatalf("expected %s=%v, got %v", key, value, payload[key])
		}
	}
}

func loadLogFixture(t *testing.T) map[string]any {
	t.Helper()

	path := filepath.Join("..", "..", "..", "contracts", "fixtures", "logs", "order-created.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var fixture map[string]any
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	return fixture
}
