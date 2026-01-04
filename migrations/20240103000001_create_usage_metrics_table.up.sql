-- Migration: create_usage_metrics_table
-- Description: Creates the usage_metrics table
-- Author: Engineering Team
-- Date: 2024-01-03
-- CockroachDB Version: v23.1+

-- Up migration
CREATE TABLE IF NOT EXISTS usage_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    stream_id UUID REFERENCES streams(id) ON DELETE CASCADE,
    metric_type STRING(50) NOT NULL,  -- stream_count, ingest_bytes, etc.
    metric_value DECIMAL(20, 2) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    INDEX idx_usage_tenant (tenant_id, period_start),
    INDEX idx_usage_stream (stream_id, period_start)
);

