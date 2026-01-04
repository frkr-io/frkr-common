package models

import "time"

// Tenant represents a tenant/organization
type Tenant struct {
	ID        string
	Name      string
	Plan      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Stream represents a message stream
type Stream struct {
	ID           string
	TenantID     string
	Name         string
	Description  string
	Status       string
	RetentionDays int
	RedpandaTopic string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

