package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"sort"
	"strconv"
	"strings"

	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func getLatestMigrationVersion(t *testing.T, path string) uint {
	files, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	var versions []uint
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		// Migration files are named like 20240101000001_create_tenants_table.up.sql
		name := f.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		parts := strings.Split(name, "_")
		if len(parts) < 1 {
			continue
		}
		version, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			t.Errorf("Failed to parse version from filename %s: %v", name, err)
			continue
		}
		versions = append(versions, uint(version))
	}

	if len(versions) == 0 {
		t.Fatalf("No migration files found in %s", path)
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})

	return versions[0]
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

	// Run migrations
	err = RunMigrations(dbURL)
	if err != nil {
		t.Fatalf("RunMigrations() failed: %v", err)
	}

	// Verify version tracking
	version, dirty, err := GetVersion(dbURL)
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
	err = RunMigrations(dbURL)
	if err != nil {
		t.Fatalf("RunMigrations() should be idempotent but failed on second run: %v", err)
	}

	// Verify version is still correct
	version2, dirty2, err := GetVersion(dbURL)
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

	// Get absolute path to migrations directory for verification
	migrationsPath, err := filepath.Abs("../migrations")
	if err != nil {
		t.Fatalf("Failed to get absolute path to migrations: %v", err)
	}

	// Run migrations - should apply all migrations in order
	err = RunMigrations(dbURL)
	if err != nil {
		t.Fatalf("RunMigrations() failed: %v", err)
	}

	// Verify final version matches the last migration
	version, dirty, err := GetVersion(dbURL)
	if err != nil {
		t.Fatalf("GetVersion() failed: %v", err)
	}
	if dirty {
		t.Error("Database should not be in dirty state")
	}
	// Verify final version matches the last migration
	expectedVersion := getLatestMigrationVersion(t, migrationsPath)
	if version != expectedVersion {
		t.Errorf("Expected version %d, got %d", expectedVersion, version)
	}
}
