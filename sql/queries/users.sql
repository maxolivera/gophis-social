-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, email, password, first_name, last_name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 AND is_deleted = false;

-- name: SoftDeleteUserByID :one
UPDATE users
SET is_deleted = true
WHERE id = $1
RETURNING is_deleted;

-- name: UpdateUser :one
UPDATE users
SET
	updated_at = $1,
	username = coalesce(sqlc.narg('username'), username),
	email = coalesce(sqlc.narg('email'), email),
	first_name = coalesce(sqlc.narg('first_name'), first_name),
	last_name = coalesce(sqlc.narg('last_name'), last_name),
	password = coalesce(sqlc.narg('password'), password)
WHERE id = $2 AND is_deleted = false
RETURNING *;
