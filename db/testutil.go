package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/frkr-io/frkr-common/migrate"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

// SetupTestDB creates a test database container (PostgreSQL-compatible), runs migrations, and returns a database connection.
// This is a shared utility for setting up test databases across frkr projects.
// The migrationsPath should be relative to the frkr-common root (e.g., "migrations").
// Note: Currently uses CockroachDB via testcontainers for testing, but works with any PostgreSQL-compatible database.
func SetupTestDB(t *testing.T, migrationsPath string) (*sql.DB, string) {
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

	// Get absolute path to migrations directory
	// migrationsPath should be relative to frkr-common root
	absMigrationsPath, err := filepath.Abs(migrationsPath)
	require.NoError(t, err)

	// Run migrations
	err = migrate.RunMigrations(migrateURL, absMigrationsPath)
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

