CREATE TABLE IF NOT EXISTS loans(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    amount DECIMAL(12, 2) NOT NULL,
    action TEXT NOT NULL, -- can be 'took' or 'paid'
    daily_interest_rate DECIMAL(12, 2), --only for 'took' action
    remaining_amount DECIMAL(12, 2), --only for 'took' action
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    Version INTEGER NOT NULL DEFAULT 1
);

ALTER TABLE loans ADD CONSTRAINT amount_check CHECK(amount > 0);

ALTER TABLE loans ADD CONSTRAINT remaining_amount_check CHECK(remaining_amount >= 0);
