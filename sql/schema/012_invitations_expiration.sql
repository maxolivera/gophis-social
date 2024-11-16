-- +goose Up
ALTER TABLE
	user_invitations
ADD COLUMN IF NOT EXISTS
	expires_at TIMESTAMP NOT NULL;

-- +goose Down
ALTER TABLE
	user_invitations
DROP COLUMN IF EXISTS
	expires_at;
