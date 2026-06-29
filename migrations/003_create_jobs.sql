CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL DEFAULT 'deploy',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payload JSONB,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_deployment ON jobs (deployment_id);
CREATE INDEX idx_jobs_status ON jobs (status);
