-- name: GetWalletByID :one
SELECT id, user_id, bank_account_id, currency, balance, created_at, updated_at
FROM wallets
WHERE id = $1;

-- name: GetWalletByUserAndCurrency :one
SELECT id, user_id, bank_account_id, currency, balance, created_at, updated_at
FROM wallets
WHERE user_id = $1 AND currency = $2;

-- name: GetWalletByUserAndCurrencyForUpdate :one
SELECT id, user_id, bank_account_id, currency, balance, created_at, updated_at
FROM wallets
WHERE user_id = $1 AND currency = $2
FOR UPDATE;

-- name: GetWalletByIDForUpdate :one
SELECT id, user_id, bank_account_id, currency, balance, created_at, updated_at
FROM wallets
WHERE id = $1
FOR UPDATE;

-- name: GetWalletByBankAccount :one
SELECT id, user_id, bank_account_id, currency, balance, created_at, updated_at
FROM wallets
WHERE bank_account_id = $1 AND bank_account_id IS NOT NULL;

-- name: CreateWallet :one
INSERT INTO wallets (id, user_id, bank_account_id, currency, balance)
VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
RETURNING *;

-- name: UpdateWalletBalance :exec
UPDATE wallets
SET balance = $1, updated_at = NOW()
WHERE id = $2;

-- name: GetUserWalletsWithBankAccounts :many
SELECT 
    w.id,
    w.user_id,
    w.bank_account_id,
    w.currency,
    w.balance,
    w.created_at,
    w.updated_at,
    ba.account_number,
    ba.bank_name,
    ba.bank_code,
    ba.account_name,
    ba.provider
FROM wallets w
LEFT JOIN bank_accounts ba ON w.bank_account_id = ba.id
WHERE w.user_id = $1
ORDER BY w.currency, w.created_at;
