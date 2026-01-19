package db

import (
	"database/sql"
	"fmt"

	"github.com/frkr-io/frkr-common/models"
)

// CreateOrGetTenant creates a tenant or returns existing one
func CreateOrGetTenant(db *sql.DB, name string) (*models.Tenant, error) {
	if name == "" {
		return nil, fmt.Errorf("tenant name cannot be empty")
	}
	if len(name) > 100 {
		return nil, fmt.Errorf("tenant name cannot exceed 100 characters")
	}

	var tenant models.Tenant
	
	// Try to get existing tenant
	err := db.QueryRow(`
		SELECT id, name, plan, created_at, updated_at, deleted_at
		FROM tenants
		WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Plan,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.DeletedAt,
	)
	
	if err == nil {
		return &tenant, nil
	}
	
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query tenant: %w", err)
	}
	
	// Create new tenant
	err = db.QueryRow(`
		INSERT INTO tenants (name, plan)
		VALUES ($1, 'free')
		RETURNING id, name, plan, created_at, updated_at, deleted_at
	`, name).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Plan,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.DeletedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	
	return &tenant, nil
}

// GetTenantByID retrieves a tenant by ID
func GetTenantByID(db *sql.DB, id string) (*models.Tenant, error) {
	if id == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}

	var tenant models.Tenant
	err := db.QueryRow(`
		SELECT id, name, plan, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Plan,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query tenant: %w", err)
	}

	return &tenant, nil
}
