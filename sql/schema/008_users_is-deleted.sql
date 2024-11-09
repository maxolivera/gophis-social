-- +goose Up
ALTER TABLE users ADD COLUMN is_deleted bool;
UPDATE users SET is_deleted = false WHERE is_deleted IS NULL;
ALTER TABLE users ALTER COLUMN is_deleted SET NOT NULL;
ALTER TABLE users ALTER COLUMN is_deleted SET DEFAULT false;

-- +goose Down
ALTER TABLE users DROP COLUMN is_deleted;
