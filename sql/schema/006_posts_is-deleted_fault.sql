-- +goose Up
UPDATE posts SET is_deleted = false WHERE is_deleted IS NULL;
ALTER TABLE posts ALTER COLUMN is_deleted SET NOT NULL;
ALTER TABLE posts ALTER COLUMN is_deleted SET DEFAULT false;

-- +goose Down
ALTER TABLE posts ALTER COLUMN is_deleted DROP NOT NULL;
ALTER TABLE posts ALTER COLUMN is_deleted DROP DEFAULT;
