-- name: CreatePosts :one
INSERT INTO posts (id, created_at, updated_at, user_id, title, content)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at;

-- name: GetPostsByUser :many
SELECT * FROM posts WHERE user_id == $1;

-- name: GetPostsById :one
SELECT * FROM posts WHERE id == $1;
