-- +goose Up

-- The pg_trgm module provides functions and
-- operators for determining the similarity of
-- ASCII alphanumeric text based on trigram matching,
-- as well as index operator classes that support
-- fast searching for similar strings.
-- https://niallburkley.com/blog/index-columns-for-like-in-postgres/
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_comments_content ON comments USING gin (content gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_posts_title ON posts USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin (tags);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts (user_id);
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments (user_id);

-- +goose Down

DROP INDEX IF EXISTS idx_comments_content;

DROP INDEX IF EXISTS idx_posts_title;
DROP INDEX IF EXISTS idx_posts_tags;

DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_posts_user_id;
DROP INDEX IF EXISTS idx_comments_post_id;

DROP EXTENSION IF EXISTS pg_trgm;
