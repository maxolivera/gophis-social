-- name: CreatePosts :one
INSERT INTO posts (id, created_at, updated_at, user_id, title, content, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, title;

-- name: GetPostByUser :many
SELECT * FROM posts WHERE user_id = $1;

-- name: GetPostById :one
SELECT * FROM posts WHERE id = $1;
