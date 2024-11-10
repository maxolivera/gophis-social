-- name: GetUserFeed :many
SELECT
	p.id, p.title, p.content, p.created_at, p.tags,
	author.id, author.username, COUNT(c.id) AS comment_count
FROM posts p
LEFT JOIN comments c ON c.post_id = p.id
LEFT JOIN users author ON p.user_id = author.id
JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
WHERE f.user_id = $1 OR p.user_id = $1
GROUP BY p.id, author.id, author.username
ORDER BY
	CASE
		WHEN NOT @sort::boolean THEN p.created_at END ASC,
	CASE
		WHEN @sort::boolean THEN p.created_at END DESC
LIMIT $2 OFFSET $3;

