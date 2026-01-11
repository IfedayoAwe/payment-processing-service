-- name: GetBankAccountByID :one
SELECT id, user_id, bank_name, bank_code, account_number, account_name, currency, provider, created_at, updated_at
FROM bank_accounts
WHERE id = $1;

-- name: GetBankAccountByAccountAndBankCode :one
SELECT id, user_id, bank_name, bank_code, account_number, account_name, currency, provider, created_at, updated_at
FROM bank_accounts
WHERE account_number = $1 AND bank_code = $2;
