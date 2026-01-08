CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    stream_id UUID REFERENCES streams(id) ON DELETE SET NULL,
    client_id STRING(255) NOT NULL,
    client_secret STRING(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, client_id),
    INDEX idx_clients_tenant (tenant_id),
    INDEX idx_clients_stream (stream_id) WHERE stream_id IS NOT NULL,
    INDEX idx_clients_deleted (deleted_at) WHERE deleted_at IS NULL
);
