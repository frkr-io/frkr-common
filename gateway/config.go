package gateway

import (
	"log"
	"os"
	"strconv"
)

// Config holds common gateway configuration
type Config struct {
	HTTPPort  int
	DBURL     string
	BrokerURL string
}

// LoadConfig loads configuration from environment variables (12-factor app pattern)
// Environment variables override any values already set in the config.
// Required fields must be set either in config or via environment variables.
func LoadConfig(cfg *Config) {
	if envURL := os.Getenv("DB_URL"); envURL != "" {
		cfg.DBURL = envURL
	}
	if envURL := os.Getenv("BROKER_URL"); envURL != "" {
		cfg.BrokerURL = envURL
	}
	if envPort := os.Getenv("HTTP_PORT"); envPort != "" {
		if port, err := strconv.Atoi(envPort); err == nil {
			cfg.HTTPPort = port
		}
	}
}

// ValidateConfig validates that all required configuration is present.
// Logs fatal error and exits if required config is missing.
func ValidateConfig(cfg *Config) {
	if cfg.DBURL == "" {
		log.Fatal("DB_URL is required (set via --db-url flag or DB_URL environment variable)")
	}
	if cfg.BrokerURL == "" {
		log.Fatal("BROKER_URL is required (set via --broker-url flag or BROKER_URL environment variable)")
	}
}

// MustLoadConfig loads and validates configuration, exiting on error
func MustLoadConfig(cfg *Config) {
	LoadConfig(cfg)
	ValidateConfig(cfg)
}
