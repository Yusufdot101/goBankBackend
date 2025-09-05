CREATE TABLE IF NOT EXISTS loan_requests(
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id BIGINT REFERENCES users,
    amount DECIMAL(12, 2) NOT NULL,
    daily_interest_rate DECIMAL(12, 2) NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING'
);

ALTER TABLE loan_requests ADD CONSTRAINT amount_check CHECK(amount > 0);
