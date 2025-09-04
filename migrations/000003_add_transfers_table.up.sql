CREATE TABLE IF NOT EXISTS transfers (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    from_user_id BIGINT  REFERENCES users NOT NULL,
    to_user_id BIGINT  REFERENCES users NOT NULL,
    amount DECIMAL(12, 2) NOT NULL
);

ALTER TABLE transfers ADD CONSTRAINT amount_check CHECK(amount > 0);
