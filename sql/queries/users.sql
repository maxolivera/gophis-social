-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, email, password)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: CreateInvitation :exec
INSERT INTO user_invitations (token, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: GetInvitation :one
SELECT user_id
FROM user_invitations
WHERE token = $1 AND expires_at > $2;

-- name: ActivateUser :exec
UPDATE users
SET is_active = true
WHERE id = $1;

-- name: DeleteToken :exec
DELETE FROM user_invitations
WHERE token = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
	AND is_deleted = false
	AND is_active = true;

-- name: GetUserByUsername :one
SELECT
	u.*, r.level, r.name
FROM users u
JOIN roles r ON u.role_id = r.id
WHERE u.username = $1
	AND is_deleted = false
	AND is_active = true;

-- name: SoftDeleteUserByID :one
UPDATE users
SET is_deleted = true
WHERE id = $1
RETURNING is_deleted;

-- name: HardDeleteUserByID :exec
DELETE FROM users
WHERE id = $1;

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
