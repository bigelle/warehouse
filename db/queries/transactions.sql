-- name: CreateNewTransaction :one
INSERT INTO transactions (user_id, item_id, amount, status, reason)
VALUES ($1, $2, $3, $4, $5)
RETURNING (id, created_at);

-- name: GetTransactionsForItem :many
SELECT *
FROM transactions
WHERE item_id = $1
LIMIT $2 OFFSET $3;
