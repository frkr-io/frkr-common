package migrate

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs all pending migrations
func RunMigrations(dbURL string, migrationsPath string) error {
	// Open database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create CockroachDB driver
	driver, err := cockroachdb.WithInstance(db, &cockroachdb.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}

	// Create file source
	source, err := (&file.File{}).Open(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to open migrations: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance(
		"file",
		source,
		"cockroachdb",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// GetVersion returns the current migration version
func GetVersion(dbURL string, migrationsPath string) (uint, bool, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return 0, false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	driver, err := cockroachdb.WithInstance(db, &cockroachdb.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create driver: %w", err)
	}

	source, err := (&file.File{}).Open(migrationsPath)
	if err != nil {
		return 0, false, fmt.Errorf("failed to open migrations: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"file",
		source,
		"cockroachdb",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	return version, dirty, nil
}

