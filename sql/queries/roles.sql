-- name: GetRoles :many
SELECT * FROM roles;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;
