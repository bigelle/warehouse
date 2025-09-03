-- name: CreateUser :one
INSERT INTO users (username, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id, username, role, created_at;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, role, created_at
FROM users
WHERE username = $1;

-- name: GetRefreshToken :one
SELECT id, role, refresh_token
FROM users
WHERE id = $1;

-- name: SetRefreshToken :one
UPDATE users
SET refresh_token = $1
WHERE id = $2
RETURNING id;

-- name: GetUserRole :one
SELECT id, role
FROM users
WHERE id = $1;
