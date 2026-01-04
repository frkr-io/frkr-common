-- Migration: create_tenants_table
-- Description: Creates the tenants table
-- Author: Engineering Team
-- Date: 2024-01-01
-- CockroachDB Version: v23.1+

-- Up migration
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name STRING(255) NOT NULL,
    plan STRING(50) NOT NULL DEFAULT 'free',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    INDEX idx_tenants_deleted (deleted_at) WHERE deleted_at IS NULL
);

