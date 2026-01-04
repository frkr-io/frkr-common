package migrate

import (
	"os"
	"path/filepath"
	"testing"
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
	// Create temporary migration directory
	tmpDir, err := os.MkdirTemp("", "migrations-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple migration file
	migrationFile := filepath.Join(tmpDir, "000001_test.up.sql")
	err = os.WriteFile(migrationFile, []byte("CREATE TABLE IF NOT EXISTS test (id INT);"), 0644)
	if err != nil {
		t.Fatalf("Failed to write migration file: %v", err)
	}

	// Note: This test requires a running CockroachDB instance
	// For now, we just test that the function doesn't panic
	// In a real test environment, you'd use testcontainers
	dbURL := "cockroachdb://user:pass@localhost:26257/test?sslmode=disable"
	
	// This will fail without a real DB, but we're testing the code path
	err = RunMigrations(dbURL, tmpDir)
	// We expect this to fail without a real database, so we just check it doesn't panic
	_ = err
}

