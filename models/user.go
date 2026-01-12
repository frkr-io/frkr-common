package models

import (
	"database/sql"
)

// TenantUser represents a user in the system
type TenantUser struct {
	ID           string       `json:"id"`
	TenantID     string       `json:"tenant_id"`
	Username     string       `json:"username"`
	PasswordHash string       `json:"-"`
	CreatedAt    sql.NullTime `json:"created_at"`
	UpdatedAt    sql.NullTime `json:"updated_at"`
	DeletedAt    *sql.NullTime `json:"deleted_at,omitempty"`
}
