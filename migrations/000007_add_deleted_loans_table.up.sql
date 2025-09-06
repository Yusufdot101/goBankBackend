CREATE TABLE IF NOT EXISTS deleted_loans (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    loan_created_at TIMESTAMPTZ NOT NULL,
    loan_last_updated_at TIMESTAMPTZ NOT NULL,
    loan_id BIGINT NOT NULL, -- the loan that was deleted/canceled
    debtor_id BIGINT REFERENCES users, -- the borrower who took the loan
    deleted_by_id BIGINT REFERENCES users, -- the admin/user who canceled/forgave it

    amount DECIMAL(12, 2) NOT NULL,
    remaining_amount DECIMAL(12, 2) NOT NULL,
    daily_interest_rate DECIMAL(12, 2) NOT NULL,
    reason TEXT NOT NULL
);
