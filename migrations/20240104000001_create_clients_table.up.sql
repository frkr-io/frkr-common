CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    stream_id UUID REFERENCES streams(id) ON DELETE SET NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, client_id)
);

CREATE INDEX IF NOT EXISTS idx_clients_tenant ON clients (tenant_id);
CREATE INDEX IF NOT EXISTS idx_clients_stream ON clients (stream_id) WHERE stream_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_clients_deleted ON clients (deleted_at) WHERE deleted_at IS NULL;
