package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/frkr-io/frkr-common/migrate"
	_ "github.com/lib/pq" // PostgreSQL driver registration
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

// SetupTestDB creates a test database container, runs migrations, and returns a database connection.
func SetupTestDB(t *testing.T) (*sql.DB, string) {
	ctx := context.Background()

	// Start CockroachDB container
	cockroachContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest",
		cockroachdb.WithInsecure(),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, cockroachContainer.Terminate(ctx))
	})

	// Get connection config
	connConfig, err := cockroachContainer.ConnectionConfig(ctx)
	require.NoError(t, err)

	// Get the mapped port
	port, err := cockroachContainer.MappedPort(ctx, "26257")
	require.NoError(t, err)

	// Build connection string for migrations (cockroachdb:// format)
	migrateURL := fmt.Sprintf("cockroachdb://%s@%s:%s/%s?sslmode=disable",
		connConfig.User,
		"localhost",
		port.Port(),
		connConfig.Database,
	)

	// Run migrations
	err = migrate.RunMigrations(migrateURL)
	require.NoError(t, err)

	// Build connection string for sql.Open (postgres:// format for lib/pq)
	dbURL := fmt.Sprintf("postgres://%s@localhost:%s/%s?sslmode=disable",
		connConfig.User,
		port.Port(),
		connConfig.Database,
	)

	// Open database connection
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0) // No limit

	// Test connection
	err = db.Ping()
	require.NoError(t, err)

	return db, dbURL
}

