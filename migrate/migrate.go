package migrate

import (
	"fmt"

	"github.com/frkr-io/frkr-common/migrations"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb" // CockroachDB driver registration
	_ "github.com/golang-migrate/migrate/v4/database/postgres"    // PostgreSQL driver registration
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs" // iofs source driver registration
)

// RunMigrations runs all pending migrations
func RunMigrations(dbURL string) error {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithSourceInstance(
		"iofs",
		d,
		dbURL,
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
func GetVersion(dbURL string) (uint, bool, error) {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return 0, false, fmt.Errorf("failed to create iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		d,
		dbURL,
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
