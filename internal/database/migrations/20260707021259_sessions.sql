-- +goose Up
CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE native_accounts (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(60) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS native_accounts;
DROP INDEX IF EXISTS sessions_expiry_idx;
DROP TABLE IF EXISTS sessions; 