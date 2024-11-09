-- +goose Up
CREATE TABLE IF NOT EXISTS followers (
	user_id UUID NOT NULL,
	follower_id UUID NOT NULL,
	created_at TIMESTAMP NOT NULL,

	PRIMARY KEY(user_id, follower_id),
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY(follower_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS followers;
