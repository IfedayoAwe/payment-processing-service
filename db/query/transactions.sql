-- name: GetTransactionByID :one
SELECT id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency,
       status, provider_name, provider_reference, exchange_rate, failure_reason,
       created_at, updated_at
FROM transactions
WHERE id = $1;

-- name: GetTransactionByIdempotencyKey :one
SELECT id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency,
       status, provider_name, provider_reference, exchange_rate, failure_reason,
       created_at, updated_at
FROM transactions
WHERE idempotency_key = $1;

-- name: CreateTransaction :one
INSERT INTO transactions (
    id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency, status, exchange_rate
)
VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateTransactionStatus :exec
UPDATE transactions
SET status = $1, updated_at = NOW()
WHERE id = $2;

-- name: UpdateTransactionWithProvider :exec
UPDATE transactions
SET provider_name = $1, provider_reference = $2, status = $3, updated_at = NOW()
WHERE id = $4;

-- name: UpdateTransactionFailure :exec
UPDATE transactions
SET status = $1, failure_reason = $2, updated_at = NOW()
WHERE id = $3;

-- name: ListTransactionsByUser :many
SELECT t.id, t.idempotency_key, t.from_wallet_id, t.to_wallet_id, t.type, t.amount, t.currency,
       t.status, t.provider_name, t.provider_reference, t.exchange_rate, t.failure_reason,
       t.created_at, t.updated_at
FROM transactions t
WHERE t.from_wallet_id IN (
    SELECT id FROM wallets WHERE user_id = $1
)
AND (
    $2::timestamp = '1970-01-01 00:00:00+00'::timestamp OR 
    (t.created_at < $2::timestamp OR (t.created_at = $2::timestamp AND t.id < $3))
)
ORDER BY t.created_at DESC, t.id DESC
LIMIT $4;
