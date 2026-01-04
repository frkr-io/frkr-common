package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/frkr-io/frkr-common/migrate"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func setupTestDB(t *testing.T) (*sql.DB, string) {
	ctx := context.Background()

	// Start CockroachDB container
	cockroachContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest",
		cockroachdb.WithInsecure(),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, cockroachContainer.Terminate(ctx))
	})

	// Get connection config (like the migrate_test.go example)
	connConfig, err := cockroachContainer.ConnectionConfig(ctx)
	require.NoError(t, err)

	// Get the mapped port (container exposes 26257, we need the host port)
	port, err := cockroachContainer.MappedPort(ctx, "26257")
	require.NoError(t, err)

	// Build connection string for migrations (cockroachdb:// format)
	migrateURL := fmt.Sprintf("cockroachdb://%s@%s:%s/%s?sslmode=disable",
		connConfig.User,
		"localhost",
		port.Port(),
		connConfig.Database,
	)

	// Get absolute path to migrations directory (from db/ directory, go up to frkr-common root)
	migrationsPath, err := filepath.Abs("../migrations")
	require.NoError(t, err)

	// Run migrations
	err = migrate.RunMigrations(migrateURL, migrationsPath)
	require.NoError(t, err)

	// Build connection string for sql.Open (postgres:// format for lib/pq)
	// lib/pq uses postgres:// format, not cockroachdb://
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

	// Test connection
	err = db.Ping()
	require.NoError(t, err)

	return db, dbURL
}

func TestCreateOrGetTenant(t *testing.T) {
	db, _ := setupTestDB(t)

	t.Run("create new tenant", func(t *testing.T) {
		tenant, err := CreateOrGetTenant(db, "test-tenant")
		require.NoError(t, err)
		require.NotEmpty(t, tenant.ID)
		require.Equal(t, "test-tenant", tenant.Name)
		require.Equal(t, "free", tenant.Plan)
		require.False(t, tenant.CreatedAt.IsZero())
	})

	t.Run("get existing tenant", func(t *testing.T) {
		// Create tenant first
		tenant1, err := CreateOrGetTenant(db, "existing-tenant")
		require.NoError(t, err)

		// Get it again
		tenant2, err := CreateOrGetTenant(db, "existing-tenant")
		require.NoError(t, err)

		// Should be the same tenant
		require.Equal(t, tenant1.ID, tenant2.ID)
		require.Equal(t, tenant1.Name, tenant2.Name)
	})

	t.Run("create multiple tenants", func(t *testing.T) {
		tenant1, err := CreateOrGetTenant(db, "tenant-1")
		require.NoError(t, err)

		tenant2, err := CreateOrGetTenant(db, "tenant-2")
		require.NoError(t, err)

		require.NotEqual(t, tenant1.ID, tenant2.ID)
	})
}

func TestCreateStream(t *testing.T) {
	db, _ := setupTestDB(t)

	// Create a tenant first
	tenant, err := CreateOrGetTenant(db, "stream-test-tenant")
	require.NoError(t, err)

	t.Run("create stream", func(t *testing.T) {
		stream, err := CreateStream(db, tenant.ID, "my-api", "Test API stream", 7)
		require.NoError(t, err)
		require.NotEmpty(t, stream.ID)
		require.Equal(t, tenant.ID, stream.TenantID)
		require.Equal(t, "my-api", stream.Name)
		require.Equal(t, "Test API stream", stream.Description)
		require.Equal(t, "active", stream.Status)
		require.Equal(t, 7, stream.RetentionDays)
		require.NotEmpty(t, stream.Topic)
		require.Contains(t, stream.Topic, "stream-")
		require.Contains(t, stream.Topic, "my-api")
	})

	t.Run("duplicate stream name fails", func(t *testing.T) {
		_, err := CreateStream(db, tenant.ID, "my-api", "Duplicate", 7)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})

	t.Run("stream topic name sanitization", func(t *testing.T) {
		stream, err := CreateStream(db, tenant.ID, "My Test API", "With spaces", 14)
		require.NoError(t, err)
		require.NotEmpty(t, stream.Topic)
		// Topic should be lowercase and use hyphens
		require.NotContains(t, stream.Topic, " ")
		require.NotContains(t, stream.Topic, "My")
	})
}

func TestGetStream(t *testing.T) {
	db, _ := setupTestDB(t)

	tenant, err := CreateOrGetTenant(db, "get-stream-tenant")
	require.NoError(t, err)

	stream, err := CreateStream(db, tenant.ID, "test-stream", "Test", 7)
	require.NoError(t, err)

	t.Run("get stream by name", func(t *testing.T) {
		found, err := GetStream(db, tenant.ID, "test-stream")
		require.NoError(t, err)
		require.Equal(t, stream.ID, found.ID)
		require.Equal(t, stream.Name, found.Name)
	})

	t.Run("get stream by ID", func(t *testing.T) {
		found, err := GetStream(db, tenant.ID, stream.ID)
		require.NoError(t, err)
		require.Equal(t, stream.ID, found.ID)
		require.Equal(t, stream.Name, found.Name)
	})

	t.Run("stream not found", func(t *testing.T) {
		_, err := GetStream(db, tenant.ID, "nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
}

func TestListStreams(t *testing.T) {
	db, _ := setupTestDB(t)

	tenant, err := CreateOrGetTenant(db, "list-streams-tenant")
	require.NoError(t, err)

	// Create multiple streams
	stream1, err := CreateStream(db, tenant.ID, "stream-1", "First", 7)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	stream2, err := CreateStream(db, tenant.ID, "stream-2", "Second", 14)
	require.NoError(t, err)

	t.Run("list all streams", func(t *testing.T) {
		streams, err := ListStreams(db, tenant.ID)
		require.NoError(t, err)
		require.Len(t, streams, 2)

		// Should be ordered by created_at DESC (newest first)
		require.Equal(t, stream2.ID, streams[0].ID)
		require.Equal(t, stream1.ID, streams[1].ID)
	})

	t.Run("list streams for different tenant", func(t *testing.T) {
		otherTenant, err := CreateOrGetTenant(db, "other-tenant")
		require.NoError(t, err)

		streams, err := ListStreams(db, otherTenant.ID)
		require.NoError(t, err)
		require.Len(t, streams, 0)
	})
}

func TestStreamOperations_Integration(t *testing.T) {
	db, _ := setupTestDB(t)

	// Full integration test: create tenant, create stream, get stream, list streams
	tenant, err := CreateOrGetTenant(db, "integration-test")
	require.NoError(t, err)
	require.NotEmpty(t, tenant.ID)

	stream, err := CreateStream(db, tenant.ID, "integration-api", "Integration test stream", 30)
	require.NoError(t, err)
	require.NotEmpty(t, stream.ID)
	require.Equal(t, "integration-api", stream.Name)
	require.Equal(t, tenant.ID, stream.TenantID)

	// Get the stream back
	retrieved, err := GetStream(db, tenant.ID, stream.Name)
	require.NoError(t, err)
	require.Equal(t, stream.ID, retrieved.ID)

	// List streams
	streams, err := ListStreams(db, tenant.ID)
	require.NoError(t, err)
	require.Len(t, streams, 1)
	require.Equal(t, stream.ID, streams[0].ID)
}

