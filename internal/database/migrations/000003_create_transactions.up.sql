CREATE TABLE IF NOT EXISTS transactions (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id UUID        REFERENCES users (id) ON DELETE SET NULL,
    to_user_id   UUID        REFERENCES users (id) ON DELETE SET NULL,
    amount       NUMERIC(20, 4) NOT NULL CHECK (amount > 0),
    type         VARCHAR(50) NOT NULL
                 CHECK (type IN ('deposit', 'withdrawal', 'transfer')),
    status       VARCHAR(20) NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'completed', 'failed', 'reversed')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_from_user ON transactions (from_user_id);
CREATE INDEX idx_transactions_to_user   ON transactions (to_user_id);
CREATE INDEX idx_transactions_status    ON transactions (status);
CREATE INDEX idx_transactions_created   ON transactions (created_at DESC);
