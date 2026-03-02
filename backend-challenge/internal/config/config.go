// Package config provides application configuration loaded from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Port           int
	MCPPort        int
	APIKey         string
	DatabaseURL    string
	OTLPEndpoint   string
	ServiceName    string
	CouponFileURLs []string // S3 URLs for gzipped coupon data files
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	// Load .env file if it exists (best effort for local dev)
	_ = godotenv.Load()

	return &Config{
		Port:         getEnvInt("PORT", 8080),
		MCPPort:      getEnvInt("MCP_PORT", 8081),
		APIKey:       getEnv("API_KEY", "apitest"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/orderdb?sslmode=disable"),
		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
		ServiceName:  getEnv("OTEL_SERVICE_NAME", "order-food-api"),
		CouponFileURLs: getEnvList("COUPON_FILE_URLS",
			"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz",
			"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz",
			"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz",
		),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

// getEnvList reads a comma-separated env var, or returns the default values.
func getEnvList(key string, defaults ...string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaults
}
