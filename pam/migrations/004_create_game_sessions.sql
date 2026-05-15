CREATE TABLE IF NOT EXISTS game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    game_id VARCHAR(255) NOT NULL,
    provider VARCHAR(100),
    started_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP,
    total_wagered DECIMAL(20,8) DEFAULT 0,
    total_won DECIMAL(20,8) DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_game_sessions_user_id ON game_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_started_at ON game_sessions(started_at DESC);
