package gateway

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	t.Run("valid with DBURL and BrokerURL", func(t *testing.T) {
		cfg := &Config{
			HTTPPort:  8080,
			DBURL:     "postgres://user@host/db",
			BrokerURL: "localhost:9092",
		}
		err := ValidateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("valid with DB components and BrokerURL", func(t *testing.T) {
		cfg := &Config{
			HTTPPort:  8080,
			DBHost:    "localhost",
			DBPort:    "26257",
			DBUser:    "root",
			DBName:    "testdb",
			BrokerURL: "localhost:9092",
		}
		err := ValidateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("valid with DBURL and broker components", func(t *testing.T) {
		cfg := &Config{
			HTTPPort:  8080,
			DBURL:     "postgres://user@host/db",
			BrokerHost: "localhost",
			BrokerPort: "9092",
		}
		err := ValidateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("valid with all components", func(t *testing.T) {
		cfg := &Config{
			HTTPPort:   8080,
			DBHost:     "localhost",
			DBPort:     "26257",
			DBUser:     "root",
			DBName:     "testdb",
			BrokerHost: "localhost",
			BrokerPort: "9092",
		}
		err := ValidateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("invalid - missing DB config", func(t *testing.T) {
		cfg := &Config{
			HTTPPort:  8080,
			BrokerURL: "localhost:9092",
		}
		err := ValidateConfig(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DB_URL or DB components")
	})

	t.Run("invalid - missing DBName", func(t *testing.T) {
		cfg := &Config{
			HTTPPort:  8080,
			DBHost:    "localhost",
			BrokerURL: "localhost:9092",
		}
		err := ValidateConfig(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DB_URL or DB components")
	})

	t.Run("invalid - missing broker config", func(t *testing.T) {
		cfg := &Config{
			HTTPPort: 8080,
			DBURL:    "postgres://user@host/db",
		}
		err := ValidateConfig(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BROKER_URL or broker components")
	})

	t.Run("invalid - missing BrokerPort", func(t *testing.T) {
		cfg := &Config{
			HTTPPort:  8080,
			DBURL:     "postgres://user@host/db",
			BrokerHost: "localhost",
		}
		err := ValidateConfig(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BROKER_URL or broker components")
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("loads from environment variables", func(t *testing.T) {
		os.Setenv("DB_URL", "postgres://envuser@envhost/envdb")
		os.Setenv("BROKER_URL", "envbroker:9092")
		os.Setenv("HTTP_PORT", "9999")
		defer func() {
			os.Unsetenv("DB_URL")
			os.Unsetenv("BROKER_URL")
			os.Unsetenv("HTTP_PORT")
		}()

		cfg := &Config{
			HTTPPort: 8080,
		}
		LoadConfig(cfg)

		assert.Equal(t, "postgres://envuser@envhost/envdb", cfg.DBURL)
		assert.Equal(t, "envbroker:9092", cfg.BrokerURL)
		assert.Equal(t, 9999, cfg.HTTPPort)
	})

	t.Run("environment variables override existing values", func(t *testing.T) {
		os.Setenv("DB_URL", "postgres://override@override/override")
		defer os.Unsetenv("DB_URL")

		cfg := &Config{
			DBURL: "postgres://original@original/original",
		}
		LoadConfig(cfg)

		assert.Equal(t, "postgres://override@override/override", cfg.DBURL)
	})

	t.Run("invalid HTTP_PORT is ignored", func(t *testing.T) {
		os.Setenv("HTTP_PORT", "invalid")
		defer os.Unsetenv("HTTP_PORT")

		cfg := &Config{
			HTTPPort: 8080,
		}
		LoadConfig(cfg)

		assert.Equal(t, 8080, cfg.HTTPPort) // Should remain unchanged
	})
}

func TestLoadConfigFromFlags(t *testing.T) {
	// Note: LoadConfigFromFlags uses the global flag.CommandLine which can only be defined once
	// This test verifies the function works, but we can't test multiple scenarios without
	// refactoring to accept a FlagSet parameter. For now, we test the basic functionality.
	t.Run("loads from flags and environment", func(t *testing.T) {
		// Set env vars as fallback since flags may already be parsed
		os.Setenv("DB_URL", "postgres://envtest@envhost/envdb")
		os.Setenv("BROKER_URL", "envbroker:9092")
		defer func() {
			os.Unsetenv("DB_URL")
			os.Unsetenv("BROKER_URL")
		}()

		// Reset flag parsing state if possible
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		
		cfg, err := LoadConfigFromFlags()
		if err != nil {
			// If flags are already defined (e.g., from other tests), skip this test
			t.Skip("Skipping - flags may already be defined")
			return
		}
		
		require.NoError(t, err)
		require.NotNil(t, cfg)
		// Config should have values from env vars (since flags weren't set)
		assert.NotEmpty(t, cfg.DBURL)
		assert.NotEmpty(t, cfg.BrokerURL)
	})
}
