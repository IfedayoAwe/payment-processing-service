-- name: CreateLedgerEntry :one
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency)
VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
RETURNING id, wallet_id, transaction_id, amount, currency, created_at;

-- name: GetWalletBalance :one
SELECT COALESCE(SUM(amount), 0)::BIGINT as balance
FROM ledger_entries
WHERE wallet_id = $1 AND currency = $2;
