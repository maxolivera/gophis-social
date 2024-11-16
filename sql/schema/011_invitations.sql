-- +goose Up
CREATE TABLE IF NOT EXISTS user_invitations (
	token bytea PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN;
UPDATE users SET is_active = false WHERE is_active IS NULL;
ALTER TABLE users ALTER COLUMN is_active SET NOT NULL;
ALTER TABLE users ALTER COLUMN is_active SET DEFAULT FALSE;

-- +goose Down
DROP TABLE IF EXISTS user_invitations;
ALTER TABLE users DROP COLUMN IF EXISTS is_active;
