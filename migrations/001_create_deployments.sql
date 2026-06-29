CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_name VARCHAR(255) NOT NULL,
    image_tag VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    params_json JSONB NOT NULL DEFAULT '{}',
    error_message TEXT,
    version VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX idx_deployments_status ON deployments (status);
CREATE INDEX idx_deployments_created_at ON deployments (created_at DESC);
