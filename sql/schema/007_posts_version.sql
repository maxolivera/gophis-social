-- +goose Up
ALTER TABLE posts ADD COLUMN version int;
UPDATE posts SET version = 0 WHERE version IS NULL;
ALTER TABLE posts ALTER COLUMN version SET NOT NULL;
ALTER TABLE posts ALTER COLUMN version SET DEFAULT 0;

-- +goose Down
ALTER TABLE posts DROP COLUMN version;
