-- name: SearchPosts :many
SELECT
    p.id, p.title, p.content, p.created_at, p.tags,
    author.id AS author_id, author.username, COUNT(c.id) AS comment_count
FROM posts p
LEFT JOIN comments c ON c.post_id = p.id
LEFT JOIN users author ON p.user_id = author.id
WHERE
    (@search::text IS NULL OR p.content ILIKE '%' || @search || '%' OR p.title ILIKE '%' || @search || '%')
    AND (@tags::text[] IS NULL OR p.tags && @tags)
    AND (@since::timestamp IS NULL OR p.created_at >= @since)
    AND (@until::timestamp IS NULL OR p.created_at <= @until)
GROUP BY p.id, author.id, author.username
ORDER BY
	CASE WHEN @sort::boolean THEN p.created_at END DESC,
	CASE WHEN NOT @sort::boolean THEN p.created_at END ASC,
	comment_count DESC
LIMIT $1 OFFSET $2;
