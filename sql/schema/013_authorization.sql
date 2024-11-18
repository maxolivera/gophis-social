-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
	id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL UNIQUE,
	level int NOT NULL DEFAULT 0,
	description TEXT NOT NULL
);

INSERT INTO
	roles (level, name, description)
VALUES
	(1, 'user', 'A user can create posts and comments'),
	(2, 'moderator', 'A moderator can update other users posts and comments'),
	(3, 'admin', 'An admin can update and delete other users posts and comments');

-- +goose Down
DROP TABLE IF EXISTS roles;
