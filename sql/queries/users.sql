-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, email, password, first_name, last_name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;
