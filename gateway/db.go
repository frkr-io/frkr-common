package gateway

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// ConnectGatewayDB creates a new database connection from GatewayBaseConfig
// Supports both connection string (DBURL) and individual components
// Connection string takes precedence if both are provided
func ConnectGatewayDB(cfg *GatewayBaseConfig) (*sql.DB, error) {
	var dbURL string

	// Prefer connection string if provided
	if cfg.DBURL != "" {
		dbURL = cfg.DBURL
	} else {
		// Build connection string from individual components
		if cfg.DBHost == "" || cfg.DBName == "" {
			return nil, fmt.Errorf("either DBURL or DBHost+DBName must be provided")
		}

		port := cfg.DBPort
		if port == "" {
			port = "26257" // Default CockroachDB port
		}

		if cfg.DBUser != "" {
			if cfg.DBPassword != "" {
				dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
					cfg.DBUser, cfg.DBPassword, cfg.DBHost, port, cfg.DBName)
			} else {
				dbURL = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable",
					cfg.DBUser, cfg.DBHost, port, cfg.DBName)
			}
		} else {
			dbURL = fmt.Sprintf("postgres://%s:%s/%s?sslmode=disable",
				cfg.DBHost, port, cfg.DBName)
		}
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
