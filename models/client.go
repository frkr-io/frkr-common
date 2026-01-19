package models

import (
	"database/sql"
)

// ClientCredential represents a client credential in the system
type ClientCredential struct {
	ID           string
	TenantID     string
	StreamID     sql.NullString
	ClientID     string
	ClientSecret string
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
	DeletedAt    *sql.NullTime
}
