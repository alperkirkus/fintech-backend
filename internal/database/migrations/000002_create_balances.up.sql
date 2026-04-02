CREATE TABLE IF NOT EXISTS balances (
    user_id         UUID        PRIMARY KEY
                    REFERENCES users (id) ON DELETE CASCADE,
    amount          NUMERIC(20, 4) NOT NULL DEFAULT 0
                    CHECK (amount >= 0),
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_balances_user_id ON balances (user_id);
