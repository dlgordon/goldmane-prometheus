package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	// GoldmaneAddr is the address of the Goldmane gRPC API
	GoldmaneAddr string

	// MetricsAddr is the address where the /metrics endpoint will be exposed
	MetricsAddr string

	// PollInterval is the interval at which to poll the Goldmane API for flow data
	PollInterval time.Duration

	// TLSEnabled indicates whether TLS should be used for the Goldmane connection
	TLSEnabled bool

	// TLSCertPath is the path to the TLS certificate file
	TLSCertPath string

	// TLSKeyPath is the path to the TLS key file
	TLSKeyPath string

	// TLSCAPath is the path to the CA certificate file
	TLSCAPath string
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	pollInterval := getEnvDuration("POLL_INTERVAL", 15*time.Second)

	return &Config{
		GoldmaneAddr: getEnv("GOLDMANE_ADDR", "localhost:9094"),
		MetricsAddr:  getEnv("METRICS_ADDR", ":9090"),
		PollInterval: pollInterval,
		TLSEnabled:   getEnvBool("TLS_ENABLED", false),
		TLSCertPath:  getEnv("TLS_CERT_PATH", ""),
		TLSKeyPath:   getEnv("TLS_KEY_PATH", ""),
		TLSCAPath:    getEnv("TLS_CA_PATH", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return time.Duration(parsed) * time.Second
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
