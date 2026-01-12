-- name: CreateLedgerEntry :one
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, account_type, balance_before, balance_after)
VALUES (gen_random_uuid()::text, $1, $2, $3, $4, 'user_wallet', $5, $6)
RETURNING id, wallet_id, transaction_id, amount, currency, account_type, balance_before, balance_after, created_at;

-- name: CreateExternalSystemCreditEntry :one
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, account_type, balance_before, balance_after)
VALUES (gen_random_uuid()::text, NULL, $1, $2, $3, 'external_wallet', $4, $5)
RETURNING id, wallet_id, transaction_id, amount, currency, account_type, balance_before, balance_after, created_at;

-- name: GetWalletBalance :one
SELECT COALESCE(SUM(amount), 0)::BIGINT as balance
FROM ledger_entries
WHERE wallet_id = $1 AND currency = $2 AND account_type = 'user_wallet';
