-- Compliance / responsible gambling table
CREATE TABLE IF NOT EXISTS compliance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    event_type VARCHAR(100) NOT NULL,  -- self_exclusion, deposit_limit, session_limit, ban
    value DECIMAL(20,8),
    period VARCHAR(50),                 -- daily, weekly, monthly
    reason TEXT,
    created_by UUID,                    -- admin user ID, NULL if player-initiated
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_compliance_user_id ON compliance_records(user_id);
