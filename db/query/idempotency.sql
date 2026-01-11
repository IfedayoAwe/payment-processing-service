-- name: CreateIdempotencyKey :one
INSERT INTO idempotency_keys (key, transaction_id)
VALUES ($1, $2)
RETURNING *;
