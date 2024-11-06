-- +goose Up

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	username TEXT NOT NULL UNIQUE,
	email citext NOT NULL UNIQUE,
	password bytea NOT NULL,
	first_name TEXT,
	last_name TEXT
);

-- +goose Down

DROP TABLE users;

DROP EXTENSION IF EXISTS citext;
