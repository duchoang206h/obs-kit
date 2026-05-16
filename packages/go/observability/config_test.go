package observability

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type configFixture struct {
	Env      map[string]string `json:"env"`
	Expected struct {
		ServiceName           string  `json:"serviceName"`
		ServiceVersion        string  `json:"serviceVersion"`
		Environment           string  `json:"environment"`
		CollectorURL          string  `json:"collectorUrl"`
		TracingSampleRate     float64 `json:"tracingSampleRate"`
		MetricsExportInterval int     `json:"metricsExportInterval"`
		LoggingLevel          string  `json:"loggingLevel"`
	} `json:"expected"`
}

func TestResolveConfigUsesEnvironment(t *testing.T) {
	fixture := loadConfigFixture(t)
	for key, value := range fixture.Env {
		t.Setenv(key, value)
	}

	config := ResolveConfig(nil)

	if config.ServiceName != fixture.Expected.ServiceName {
		t.Fatalf("expected service name orders, got %s", config.ServiceName)
	}
	if config.ServiceVersion != fixture.Expected.ServiceVersion {
		t.Fatalf("expected service version 1.2.3, got %s", config.ServiceVersion)
	}
	if config.Environment != fixture.Expected.Environment {
		t.Fatalf("expected environment test, got %s", config.Environment)
	}
	if config.Collector.URL != fixture.Expected.CollectorURL {
		t.Fatalf("expected collector URL from env, got %s", config.Collector.URL)
	}
	if config.Tracing.SampleRate == nil || *config.Tracing.SampleRate != fixture.Expected.TracingSampleRate {
		t.Fatalf("expected sample rate 0.5, got %v", config.Tracing.SampleRate)
	}
	if config.Metrics.ExportInterval != fixture.Expected.MetricsExportInterval {
		t.Fatalf("expected export interval 30000, got %d", config.Metrics.ExportInterval)
	}
	if config.Logging.Level != fixture.Expected.LoggingLevel {
		t.Fatalf("expected log level debug, got %s", config.Logging.Level)
	}
}

func TestResolveConfigMergesExplicitValuesWithEnvironment(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "env-service")
	t.Setenv("ENVIRONMENT", "test")
	t.Setenv("LOG_LEVEL", "debug")

	config := ResolveConfig(&Config{
		ServiceName: "explicit-service",
		Tracing: TracingConfig{
			Enabled: Bool(false),
		},
	})

	if config.ServiceName != "explicit-service" {
		t.Fatalf("expected explicit service name, got %s", config.ServiceName)
	}
	if config.Environment != "test" {
		t.Fatalf("expected environment from env, got %s", config.Environment)
	}
	if config.Logging.Level != "debug" {
		t.Fatalf("expected log level from env, got %s", config.Logging.Level)
	}
	if config.Tracing.Enabled == nil || *config.Tracing.Enabled {
		t.Fatalf("expected explicit tracing disabled")
	}
}

func TestResolveConfigPreservesExplicitZeroSampleRate(t *testing.T) {
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "1.0")

	config := ResolveConfig(&Config{
		Tracing: TracingConfig{
			SampleRate: Float64(0),
		},
	})

	if config.Tracing.SampleRate == nil || *config.Tracing.SampleRate != 0 {
		t.Fatalf("expected explicit zero sample rate, got %v", config.Tracing.SampleRate)
	}
}

func loadConfigFixture(t *testing.T) configFixture {
	t.Helper()

	path := filepath.Join("..", "..", "..", "contracts", "fixtures", "config", "env-basic.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var fixture configFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	return fixture
}
