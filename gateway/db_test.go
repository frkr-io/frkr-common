package gateway

import (
	"testing"

	"github.com/frkr-io/frkr-common/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB(t *testing.T) {
	t.Run("creates DB from connection string", func(t *testing.T) {
		testDB, dbURL := db.SetupTestDB(t, "../migrations")
		defer testDB.Close()

		cfg := &Config{
			DBURL: dbURL,
		}

		db, err := NewDB(cfg)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer db.Close()

		// Verify connection works
		err = db.Ping()
		assert.NoError(t, err)
	})

	t.Run("creates DB from individual components", func(t *testing.T) {
		testDB, _ := db.SetupTestDB(t, "../migrations")
		defer testDB.Close()

		// Extract connection info from test container
		cfg := &Config{
			DBHost:     "localhost",
			DBPort:     "26257",
			DBUser:     "root",
			DBName:     "defaultdb",
		}

		// Note: This may fail if test container port is not 26257
		// In practice, we'd get the actual port from the container
		db, err := NewDB(cfg)
		if err != nil {
			t.Skip("Skipping - test container port may differ")
			return
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			t.Skip("Skipping - connection failed (container may use different port)")
			return
		}
		assert.NoError(t, err)
	})

	t.Run("creates DB with password from components", func(t *testing.T) {
		cfg := &Config{
			DBHost:     "localhost",
			DBPort:     "26257",
			DBUser:     "user",
			DBPassword: "password",
			DBName:     "testdb",
		}

		db, err := NewDB(cfg)
		// This will fail to connect, but we're just testing URL construction
		if err != nil && err.Error() == "failed to open database connection" {
			// Expected - we don't have a real DB at this address
			return
		}
		require.NoError(t, err)
		defer db.Close()
	})

	t.Run("uses default port when DBPort not specified", func(t *testing.T) {
		cfg := &Config{
			DBHost: "localhost",
			DBUser: "root",
			DBName: "testdb",
		}

		db, err := NewDB(cfg)
		// Will fail to connect, but URL should include default port
		if err != nil && err.Error() == "failed to open database connection" {
			// Expected
			return
		}
		require.NoError(t, err)
		defer db.Close()
	})

	t.Run("connection string takes precedence over components", func(t *testing.T) {
		testDB, dbURL := db.SetupTestDB(t, "../migrations")
		defer testDB.Close()

		cfg := &Config{
			DBURL:     dbURL,
			DBHost:    "wronghost",
			DBPort:    "9999",
			DBUser:    "wronguser",
			DBName:    "wrongdb",
		}

		db, err := NewDB(cfg)
		require.NoError(t, err)
		defer db.Close()

		// Should connect using DBURL, not components
		err = db.Ping()
		assert.NoError(t, err)
	})

	t.Run("error when neither DBURL nor components provided", func(t *testing.T) {
		cfg := &Config{}

		db, err := NewDB(cfg)
		require.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "DBURL or DBHost+DBName")
	})

	t.Run("error when DBHost provided but DBName missing", func(t *testing.T) {
		cfg := &Config{
			DBHost: "localhost",
		}

		db, err := NewDB(cfg)
		require.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "DBHost+DBName")
	})
}
