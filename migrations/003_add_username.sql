ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(16) UNIQUE;
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
