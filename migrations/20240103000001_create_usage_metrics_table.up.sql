CREATE TABLE IF NOT EXISTS usage_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    stream_id UUID REFERENCES streams(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL,  -- stream_count, ingest_bytes, etc.
    metric_value DECIMAL(20, 2) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_usage_tenant ON usage_metrics (tenant_id, period_start);
CREATE INDEX IF NOT EXISTS idx_usage_stream ON usage_metrics (stream_id, period_start);

