CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    code TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS users_permissions (
    user_id BIGINT REFERENCES users ON DELETE CASCADE,
    permission_id BIGINT REFERENCES permissions ON DELETE CASCADE,
    PRIMARY KEY(user_id, permission_id)
);

INSERT INTO permissions (code)
VALUES
    ('APPROVE_LOANS'),
    ('DELETE_LOANS'),
    ('ADMIN'),
    ('SUPERUSER');
