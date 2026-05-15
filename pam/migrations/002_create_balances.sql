CREATE TABLE IF NOT EXISTS balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cash_balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    bonus_balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    locked_balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    version INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT positive_cash CHECK (cash_balance >= 0),
    CONSTRAINT positive_bonus CHECK (bonus_balance >= 0),
    CONSTRAINT positive_locked CHECK (locked_balance >= 0)
);
