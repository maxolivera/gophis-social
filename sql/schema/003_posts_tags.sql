-- +goose Up

ALTER TABLE posts ADD COLUMN tags text[];

-- +goose Down

ALTER TABLE posts DROP COLUMN tags;
