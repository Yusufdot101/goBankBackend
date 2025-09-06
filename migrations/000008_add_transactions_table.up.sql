CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id BIGINT NOT NULL,
    amount DECIMAL(12, 2),
    action TEXT NOT NULL,
    performed_by TEXT NOT NULL
);

ALTER TABLE transactions ADD CONSTRAINT amount_check CHECK(amount > 0);
