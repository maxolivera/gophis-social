-- name: FollowByID :exec
INSERT INTO followers(created_at, user_id, follower_id)
VALUES ($1, $2, $3);

-- name: UnfollowByID :exec
DELETE FROM followers WHERE user_id = $1 AND follower_id = $2;
