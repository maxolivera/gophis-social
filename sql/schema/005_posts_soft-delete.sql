-- +goose Up
ALTER TABLE posts ADD COLUMN IF NOT EXISTS is_deleted bool;

-- +goose Down
ALTER TABLE posts DROP COLUMN IF EXISTS is_deleted;
