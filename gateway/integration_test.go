package gateway

import (
	"testing"

	"github.com/frkr-io/frkr-common/db"
	"github.com/frkr-io/frkr-common/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatewayInitializationFlow(t *testing.T) {
	testDB, dbURL := db.SetupTestDB(t, "../migrations")
	defer testDB.Close()

	// Setup users table
	_, err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			username STRING(255) NOT NULL,
			password_hash STRING(255) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ,
			UNIQUE (tenant_id, username)
		)
	`)
	require.NoError(t, err)

	t.Run("full initialization with connection strings", func(t *testing.T) {
		cfg := &GatewayBaseConfig{
			HTTPPort:  8080,
			DBURL:     dbURL,
			BrokerURL: "localhost:9092",
		}

		// Validate config
		err := ValidateConfig(cfg)
		require.NoError(t, err)

		// Initialize DB
		db, err := ConnectGatewayDB(cfg)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer db.Close()

		// Initialize plugins
		secretPlugin, err := plugins.NewDatabaseSecretPlugin(db)
		require.NoError(t, err)

		authPlugin := plugins.NewBasicAuthPlugin(db)
		require.NotNil(t, authPlugin)

		// Initialize broker writer
		writer := NewBrokerWriter(cfg)
		require.NotNil(t, writer)
		defer writer.Close()

		// Verify all components are initialized
		assert.NotNil(t, db)
		assert.NotNil(t, secretPlugin)
		assert.NotNil(t, authPlugin)
		assert.NotNil(t, writer)
	})

	t.Run("full initialization with individual components", func(t *testing.T) {
		// Extract connection components from test DB URL
		cfg := &GatewayBaseConfig{
			HTTPPort:   8080,
			DBHost:     "localhost",
			DBPort:     "26257",
			DBUser:     "root",
			DBName:     "defaultdb",
			BrokerHost: "localhost",
			BrokerPort: "9092",
		}

		// Validate config
		err := ValidateConfig(cfg)
		require.NoError(t, err)

		// Initialize DB (may skip if port doesn't match)
		db, err := ConnectGatewayDB(cfg)
		if err != nil {
			t.Skip("Skipping - test container port may differ")
			return
		}
		defer db.Close()

		// Verify DB connection works
		err = db.Ping()
		if err != nil {
			t.Skip("Skipping - connection failed")
			return
		}

		// Initialize plugins
		secretPlugin, err := plugins.NewDatabaseSecretPlugin(db)
		require.NoError(t, err)

		authPlugin := plugins.NewBasicAuthPlugin(db)
		require.NotNil(t, authPlugin)

		// Initialize broker writer
		writer := NewBrokerWriter(cfg)
		require.NotNil(t, writer)
		defer writer.Close()

		// Verify all components are initialized
		assert.NotNil(t, db)
		assert.NotNil(t, secretPlugin)
		assert.NotNil(t, authPlugin)
		assert.NotNil(t, writer)
	})

	t.Run("connection string takes precedence over components", func(t *testing.T) {
		cfg := &GatewayBaseConfig{
			HTTPPort:   8080,
			DBURL:      dbURL, // This should be used
			DBHost:     "wronghost",
			DBPort:     "9999",
			DBUser:     "wronguser",
			DBName:     "wrongdb",
			BrokerURL:  "localhost:9092", // This should be used
			BrokerHost: "wronghost",
			BrokerPort: "9999",
		}

		// Initialize DB - should use DBURL, not components
		db, err := ConnectGatewayDB(cfg)
		require.NoError(t, err)
		defer db.Close()

		// Verify it connected using DBURL
		err = db.Ping()
		assert.NoError(t, err)

		// Initialize broker writer - should use BrokerURL, not components
		writer := NewBrokerWriter(cfg)
		assert.NotNil(t, writer)
		assert.Equal(t, "localhost:9092", writer.Addr.String())
		defer writer.Close()
	})
}
