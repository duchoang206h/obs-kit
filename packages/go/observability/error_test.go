package observability

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestErrorAttrsAddsErrorFields(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&Config{ServiceName: "orders"}, &buffer)

	logger.Error("failed", ErrorAttrs(errors.New("bad order"))...)

	var payload map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON log payload: %v", err)
	}

	if payload["error.type"] != "*errors.errorString" {
		t.Fatalf("expected error type, got %v", payload["error.type"])
	}
	if payload["error.message"] != "bad order" {
		t.Fatalf("expected error message, got %v", payload["error.message"])
	}
}
