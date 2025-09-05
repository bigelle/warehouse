-- name: GetNItemsOffset :many
SELECT *
FROM items
ORDER BY id
LIMIT $1 OFFSET $2;
