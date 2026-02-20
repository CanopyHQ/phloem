-- Knowledge Base Schema for Self-Healing System
-- This schema stores failure records, remediation records, and policy effectiveness
-- for the MAPE-K control plane

-- Failure records table
CREATE TABLE IF NOT EXISTS failure_records (
    id TEXT PRIMARY KEY,
    correlation_id TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    category TEXT NOT NULL,
    type TEXT NOT NULL,
    subtype TEXT NOT NULL,
    instance TEXT,
    severity INTEGER NOT NULL,
    impact REAL NOT NULL,
    raw_event JSONB,
    analysis JSONB,
    remediation_id TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_method TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for failure_records
CREATE INDEX IF NOT EXISTS idx_failure_records_correlation ON failure_records(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_failure_records_category_type ON failure_records(category, type, subtype);
CREATE INDEX IF NOT EXISTS idx_failure_records_timestamp ON failure_records(timestamp);
CREATE INDEX IF NOT EXISTS idx_failure_records_resolved ON failure_records(resolved_at) WHERE resolved_at IS NOT NULL;

-- Remediation records table
CREATE TABLE IF NOT EXISTS remediation_records (
    id TEXT PRIMARY KEY,
    failure_id TEXT REFERENCES failure_records(id) ON DELETE CASCADE,
    policy_id TEXT NOT NULL,
    action TEXT NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    success BOOLEAN,
    outcome JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for remediation_records
CREATE INDEX IF NOT EXISTS idx_remediation_records_failure ON remediation_records(failure_id);
CREATE INDEX IF NOT EXISTS idx_remediation_records_policy ON remediation_records(policy_id);
CREATE INDEX IF NOT EXISTS idx_remediation_records_success ON remediation_records(success) WHERE success IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_remediation_records_started ON remediation_records(started_at);

-- Policy effectiveness table
CREATE TABLE IF NOT EXISTS policy_effectiveness (
    policy_id TEXT NOT NULL,
    failure_pattern TEXT NOT NULL,
    total_attempts INTEGER DEFAULT 0,
    successful_attempts INTEGER DEFAULT 0,
    avg_resolution_time INTERVAL,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (policy_id, failure_pattern)
);

-- Index for policy_effectiveness
CREATE INDEX IF NOT EXISTS idx_policy_effectiveness_updated ON policy_effectiveness(last_updated);

-- Function to update policy effectiveness
CREATE OR REPLACE FUNCTION update_policy_effectiveness(
    p_policy_id TEXT,
    p_failure_pattern TEXT,
    p_success BOOLEAN,
    p_resolution_time INTERVAL
) RETURNS VOID AS $$
BEGIN
    INSERT INTO policy_effectiveness (policy_id, failure_pattern, total_attempts, successful_attempts, avg_resolution_time, last_updated)
    VALUES (p_policy_id, p_failure_pattern, 1, CASE WHEN p_success THEN 1 ELSE 0 END, p_resolution_time, NOW())
    ON CONFLICT (policy_id, failure_pattern) DO UPDATE
    SET
        total_attempts = policy_effectiveness.total_attempts + 1,
        successful_attempts = policy_effectiveness.successful_attempts + CASE WHEN p_success THEN 1 ELSE 0 END,
        avg_resolution_time = CASE
            WHEN policy_effectiveness.avg_resolution_time IS NULL THEN p_resolution_time
            ELSE (policy_effectiveness.avg_resolution_time * (policy_effectiveness.total_attempts - 1) + p_resolution_time) / policy_effectiveness.total_attempts
        END,
        last_updated = NOW();
END;
$$ LANGUAGE plpgsql;
