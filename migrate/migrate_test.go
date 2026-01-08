package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func TestRunMigrations_InvalidPath(t *testing.T) {
	// Test with invalid migration path
	err := RunMigrations("cockroachdb://user:pass@localhost:26257/test?sslmode=disable", "/nonexistent/path")
	if err == nil {
		t.Error("RunMigrations() should fail with invalid path")
	}
}

func TestGetVersion_InvalidPath(t *testing.T) {
	// Test with invalid migration path
	_, _, err := GetVersion("cockroachdb://user:pass@localhost:26257/test?sslmode=disable", "/nonexistent/path")
	if err == nil {
		t.Error("GetVersion() should fail with invalid path")
	}
}

func TestRunMigrations_ValidMigrations(t *testing.T) {
	ctx := context.Background()

	// Start CockroachDB container using the module (handles initialization, health checks, etc.)
	cockroachContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest",
		cockroachdb.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("Failed to start CockroachDB container: %v", err)
	}
	defer func() {
		if err := cockroachContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %v", err)
		}
	}()

	// Get connection config and build connection string
	connConfig, err := cockroachContainer.ConnectionConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection config: %v", err)
	}

	// Build cockroachdb:// connection string from config
	dbURL := fmt.Sprintf("cockroachdb://%s@%s:%d/%s?sslmode=disable",
		connConfig.User,
		connConfig.Host,
		connConfig.Port,
		connConfig.Database,
	)

	// Get absolute path to migrations directory
	migrationsPath, err := filepath.Abs("../migrations")
	if err != nil {
		t.Fatalf("Failed to get absolute path to migrations: %v", err)
	}

	// Run migrations
	err = RunMigrations(dbURL, migrationsPath)
	if err != nil {
		t.Fatalf("RunMigrations() failed: %v", err)
	}

	// Verify version tracking
	version, dirty, err := GetVersion(dbURL, migrationsPath)
	if err != nil {
		t.Fatalf("GetVersion() failed: %v", err)
	}
	if dirty {
		t.Error("Database should not be in dirty state after successful migration")
	}
	if version == 0 {
		t.Error("Version should be set after running migrations")
	}

	// Verify that running migrations again doesn't fail (should be idempotent)
	err = RunMigrations(dbURL, migrationsPath)
	if err != nil {
		t.Fatalf("RunMigrations() should be idempotent but failed on second run: %v", err)
	}

	// Verify version is still correct
	version2, dirty2, err := GetVersion(dbURL, migrationsPath)
	if err != nil {
		t.Fatalf("GetVersion() failed on second call: %v", err)
	}
	if dirty2 {
		t.Error("Database should not be in dirty state")
	}
	if version2 != version {
		t.Errorf("Version should remain the same on second run: got %d, expected %d", version2, version)
	}
}

func TestGetVersion_NoMigrations(t *testing.T) {
	ctx := context.Background()

	// Start CockroachDB container using the module
	cockroachContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest",
		cockroachdb.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("Failed to start CockroachDB container: %v", err)
	}
	defer func() {
		if err := cockroachContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %v", err)
		}
	}()

	// Get connection config and build connection string
	connConfig, err := cockroachContainer.ConnectionConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection config: %v", err)
	}

	// Build cockroachdb:// connection string from config
	dbURL := fmt.Sprintf("cockroachdb://%s@%s:%d/%s?sslmode=disable",
		connConfig.User,
		connConfig.Host,
		connConfig.Port,
		connConfig.Database,
	)

	// Create empty migrations directory
	tmpDir, err := os.MkdirTemp("", "migrations-empty")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// GetVersion should return 0, false, nil for a database with no migrations
	version, dirty, err := GetVersion(dbURL, tmpDir)
	if err != nil {
		t.Fatalf("GetVersion() should not fail on empty migrations: %v", err)
	}
	if version != 0 {
		t.Errorf("Version should be 0 for empty migrations, got %d", version)
	}
	if dirty {
		t.Error("Dirty should be false for empty migrations")
	}
}

func TestRunMigrations_MultipleMigrations(t *testing.T) {
	ctx := context.Background()

	// Start CockroachDB container using the module
	cockroachContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest",
		cockroachdb.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("Failed to start CockroachDB container: %v", err)
	}
	defer func() {
		if err := cockroachContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %v", err)
		}
	}()

	// Get connection config and build connection string
	connConfig, err := cockroachContainer.ConnectionConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection config: %v", err)
	}

	// Build cockroachdb:// connection string from config
	dbURL := fmt.Sprintf("cockroachdb://%s@%s:%d/%s?sslmode=disable",
		connConfig.User,
		connConfig.Host,
		connConfig.Port,
		connConfig.Database,
	)

	// Get absolute path to migrations directory
	migrationsPath, err := filepath.Abs("../migrations")
	if err != nil {
		t.Fatalf("Failed to get absolute path to migrations: %v", err)
	}

	// Run migrations - should apply all migrations in order
	err = RunMigrations(dbURL, migrationsPath)
	if err != nil {
		t.Fatalf("RunMigrations() failed: %v", err)
	}

	// Verify final version matches the last migration
	version, dirty, err := GetVersion(dbURL, migrationsPath)
	if err != nil {
		t.Fatalf("GetVersion() failed: %v", err)
	}
	if dirty {
		t.Error("Database should not be in dirty state")
	}
	// Should be version 20240104000001 (the last migration - clients table)
	expectedVersion := uint(20240104000001)
	if version != expectedVersion {
		t.Errorf("Expected version %d, got %d", expectedVersion, version)
	}
}
