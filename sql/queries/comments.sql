-- name: GetCommentsByPost :many
SELECT comments.post_id, comments.id, comments.content, users.username, users.email, users.first_name, comments.created_at
FROM comments
LEFT JOIN users ON comments.user_id = users.id
WHERE comments.post_id = $1
ORDER BY comments.created_at DESC;

-- name: CreateCommentInPost :exec
INSERT INTO comments (id, user_id, post_id, created_at, updated_at, content)
VALUES ($1, $2, $3, $4, $5, $6);
