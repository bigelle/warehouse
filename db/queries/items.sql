-- name: CreateItem :one
INSERT INTO items (name)
VALUES ($1)
RETURNING uuid, name, created_at;

-- name: GetNItemsOffset :many
SELECT uuid, name, quantity, created_at, updated_at
FROM items
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: GetItem :one
SELECT uuid, name, quantity, created_at, updated_at
FROM items
WHERE uuid = $1;

-- name: PatchItem :one
UPDATE items
SET
    name = COALESCE(sqlc.narg('name'), name),
    quantity = COALESCE(sqlc.narg('quantity'), quantity),
    updated_at = now()
WHERE uuid = $1 
RETURNING uuid, name, quantity, created_at, updated_at;

-- name: DeleteItem :exec
DELETE FROM items
WHERE uuid = $1;
