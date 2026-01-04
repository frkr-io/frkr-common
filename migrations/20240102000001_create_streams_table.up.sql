CREATE TABLE IF NOT EXISTS streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name STRING(255) NOT NULL,
    description TEXT,
    status STRING(50) NOT NULL DEFAULT 'active',
    retention_days INT NOT NULL DEFAULT 7,
    topic STRING(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, name),
    INDEX idx_streams_tenant (tenant_id),
    INDEX idx_streams_status (status) WHERE deleted_at IS NULL
);

