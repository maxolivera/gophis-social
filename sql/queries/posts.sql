-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, user_id, title, content, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, title;

-- name: HardDeletePostByID :one
DELETE FROM posts WHERE id = $1 and version = $2 RETURNING *;

-- name: SoftDeletePostByID :one
UPDATE posts
SET is_deleted = true
WHERE id = $1 and version = $2
RETURNING is_deleted;

-- name: UpdatePost :one
UPDATE posts
SET
	updated_at = $1,
	title = coalesce(sqlc.narg('title'), title),
	content = coalesce(sqlc.narg('content'), content),
	tags = coalesce(sqlc.narg('tags'), tags)
WHERE id = $2 AND is_deleted = false AND version = $3
RETURNING *;

-- name: GetPostByUser :many
SELECT * FROM posts WHERE user_id = $1 AND is_deleted = false;

-- name: GetPostById :one
SELECT * FROM posts WHERE id = $1 AND is_deleted = false;
