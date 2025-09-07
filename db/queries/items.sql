-- name: CreateItem :one
INSERT INTO items (name)
VALUES ($1)
RETURNING uuid, name, created_at;

-- name: GetNItemsOffset :many
SELECT *
FROM items
ORDER BY id
LIMIT $1 OFFSET $2;
