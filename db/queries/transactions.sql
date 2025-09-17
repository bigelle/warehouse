-- name: CreateNewTransaction :one
INSERT INTO transactions (user_id, item_id, type, amount, status, reason)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at;

-- name: GetTransaction :one
SELECT *
FROM transactions
WHERE id = $1;

-- name: GetAllTransactions :many
SELECT * FROM transactions
LIMIT $1 OFFSET $2;

-- name: GetTransactionsForItem :many
SELECT * FROM transactions
WHERE item_id = $1
LIMIT $2 OFFSET $3;

-- name: GetTransactionsForUser :many
SELECT * FROM transactions
WHERE user_id = $1
LIMIT $2 OFFSET $3;
