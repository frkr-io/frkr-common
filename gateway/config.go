package gateway

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// GatewayBaseConfig holds common gateway configuration
// Supports both connection strings and individual components
type GatewayBaseConfig struct {
	HTTPPort int

	// Database configuration - connection string or individual components
	DBURL     string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPassword string
	DBName    string

	// Broker configuration - connection string or individual components
	BrokerURL  string
	BrokerHost string
	BrokerPort string
}

// LoadConfig loads configuration from environment variables (12-factor app pattern)
// Environment variables override any values already set in the config.
// Required fields must be set either in config or via environment variables.
func LoadConfig(cfg *GatewayBaseConfig) {
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
// Returns error if required config is missing.
func ValidateConfig(cfg *GatewayBaseConfig) error {
	// Validate database config - need either DBURL or individual components
	hasDBURL := cfg.DBURL != ""
	hasDBComponents := cfg.DBHost != "" && cfg.DBName != ""
	if !hasDBURL && !hasDBComponents {
		return fmt.Errorf("DB_URL or DB components (DBHost, DBName) are required (set via --db-url flag or DB_URL environment variable)")
	}

	// Validate broker config - need either BrokerURL or individual components
	hasBrokerURL := cfg.BrokerURL != ""
	hasBrokerComponents := cfg.BrokerHost != "" && cfg.BrokerPort != ""
	if !hasBrokerURL && !hasBrokerComponents {
		return fmt.Errorf("BROKER_URL or broker components (BrokerHost, BrokerPort) are required (set via --broker-url flag or BROKER_URL environment variable)")
	}

	return nil
}

// MustLoadConfig loads and validates configuration, exiting on error
func MustLoadConfig(cfg *GatewayBaseConfig) {
	LoadConfig(cfg)
	if err := ValidateConfig(cfg); err != nil {
		log.Fatal(err)
	}
}

// LoadConfigFromFlags loads configuration from command-line flags and environment variables
// Defines flags, parses them if not already parsed, and returns GatewayBaseConfig
// Returns GatewayBaseConfig populated from flags and environment variables, or error if validation fails
func LoadConfigFromFlags() (*GatewayBaseConfig, error) {
	httpPort := flag.Int("http-port", 8080, "HTTP server port")
	dbURL := flag.String("db-url", "", "Postgres-compatible database connection URL (can use DB_URL env var instead)")
	brokerURL := flag.String("broker-url", "", "Broker URL (can use BROKER_URL env var instead)")

	// Parse flags if not already parsed
	if !flag.Parsed() {
		flag.Parse()
	}

	cfg := &GatewayBaseConfig{
		HTTPPort: *httpPort,
		DBURL:    *dbURL,
		BrokerURL: *brokerURL,
	}

	LoadConfig(cfg)
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
