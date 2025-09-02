-- name: CreateUser :one
INSERT INTO users (username, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id, username, role, created_at;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, role, created_at
FROM users
WHERE username = $1;

