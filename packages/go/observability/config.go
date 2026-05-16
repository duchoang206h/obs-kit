package observability

import (
	"os"
	"strconv"
)

func Bool(value bool) *bool {
	return &value
}

func Float64(value float64) *float64 {
	return &value
}

type CollectorConfig struct {
	URL string
}

type TracingConfig struct {
	Enabled     *bool
	SampleRate  *float64
	IgnorePaths []string
}

type MetricsConfig struct {
	Enabled        *bool
	ExportInterval int
}

type LoggingConfig struct {
	Level string
}

type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Collector      CollectorConfig
	Tracing        TracingConfig
	Metrics        MetricsConfig
	Logging        LoggingConfig
}

func ResolveConfig(config *Config) Config {
	envConfig := Config{
		ServiceName:    getEnv("unknown-service", "OTEL_SERVICE_NAME"),
		ServiceVersion: getEnv("unknown", "OTEL_SERVICE_VERSION"),
		Environment:    getEnv("development", "ENVIRONMENT"),
		Collector: CollectorConfig{
			URL: getEnv("http://localhost:4318", "OTEL_EXPORTER_OTLP_ENDPOINT"),
		},
		Tracing: TracingConfig{
			Enabled:     Bool(true),
			SampleRate:  Float64(getFloat("OTEL_TRACES_SAMPLER_ARG", 1.0)),
			IgnorePaths: []string{"/health", "/metrics"},
		},
		Metrics: MetricsConfig{
			Enabled:        Bool(true),
			ExportInterval: getInt("OTEL_METRICS_EXPORT_INTERVAL", 60000),
		},
		Logging: LoggingConfig{
			Level: getEnv("info", "LOG_LEVEL"),
		},
	}

	if config == nil {
		return envConfig
	}

	return Config{
		ServiceName:    coalesceString(config.ServiceName, envConfig.ServiceName),
		ServiceVersion: coalesceString(config.ServiceVersion, envConfig.ServiceVersion),
		Environment:    coalesceString(config.Environment, envConfig.Environment),
		Collector: CollectorConfig{
			URL: coalesceString(config.Collector.URL, envConfig.Collector.URL),
		},
		Tracing: TracingConfig{
			Enabled:     coalesceBool(config.Tracing.Enabled, envConfig.Tracing.Enabled),
			SampleRate:  coalesceFloat(config.Tracing.SampleRate, envConfig.Tracing.SampleRate),
			IgnorePaths: coalesceSlice(config.Tracing.IgnorePaths, envConfig.Tracing.IgnorePaths),
		},
		Metrics: MetricsConfig{
			Enabled:        coalesceBool(config.Metrics.Enabled, envConfig.Metrics.Enabled),
			ExportInterval: coalesceInt(config.Metrics.ExportInterval, envConfig.Metrics.ExportInterval),
		},
		Logging: LoggingConfig{
			Level: coalesceString(config.Logging.Level, envConfig.Logging.Level),
		},
	}
}

func coalesceString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func coalesceFloat(value *float64, fallback *float64) *float64 {
	if value == nil {
		return fallback
	}
	return value
}

func coalesceInt(value int, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func coalesceBool(value *bool, fallback *bool) *bool {
	if value == nil {
		return fallback
	}
	return value
}

func coalesceSlice(value []string, fallback []string) []string {
	if len(value) == 0 {
		return fallback
	}
	return value
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func floatValue(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

func getEnv(fallback string, keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return fallback
}

func getFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
