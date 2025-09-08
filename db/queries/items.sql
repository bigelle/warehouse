-- name: CreateItem :one
INSERT INTO items (name)
VALUES ($1)
RETURNING uuid, name, created_at;

-- name: PatchItem :one
UPDATE items
SET
    name = COALESCE($1, name),
    quantity = COALESCE($2, quantity),
    updated_at = now()
WHERE id = $3
RETURNING uuid, name, quantity, created_at, updated_at;

-- name: GetItemQuantity :one
SELECT quantity
FROM items
WHERE uuid = $1;

-- name: GetItemQuantityConcurrable :one
SELECT quantity
FROM items
WHERE uuid = $1
FOR UPDATE;

-- name: SetItemQuantity :one
UPDATE items
SET quantity = $1, updated_at = now()
WHERE uuid = $2
RETURNING uuid, name, quantity, updated_at;

-- name: GetNItemsOffset :many
SELECT uuid, name, quantity, created_at, updated_at
FROM items
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: GetItem :one
SELECT uuid, name, quantity, created_at, updated_at
FROM items
WHERE uuid = $1;
