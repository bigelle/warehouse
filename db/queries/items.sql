-- name: CreateItem :one
INSERT INTO items (name)
VALUES ($1)
RETURNING uuid, name, created_at;

-- name: SetItemQuantity :one
UPDATE items
SET quantity = $1, updated_at = now()
WHERE uuid = $2
RETURNING uuid, name, quantity, updated_at;

-- name: RestockItem :one
UPDATE items
SET quantity = quantity + $1, updated_at = now()
WHERE uuid = $2
RETURNING uuid, name, quantity, updated_at;

-- name: GetNItemsOffset :many
SELECT *
FROM items
ORDER BY id
LIMIT $1 OFFSET $2;
