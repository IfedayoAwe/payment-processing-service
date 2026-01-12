DROP INDEX IF EXISTS idx_processed_jobs_expires_at;
DROP TABLE IF EXISTS processed_jobs;

DROP INDEX IF EXISTS idx_outbox_job_type;
DROP INDEX IF EXISTS idx_outbox_processed;
DROP TABLE IF EXISTS outbox;

DROP INDEX IF EXISTS idx_webhook_events_created_at;
DROP INDEX IF EXISTS idx_webhook_events_processed;
DROP INDEX IF EXISTS idx_webhook_events_transaction_id;
DROP INDEX IF EXISTS idx_webhook_events_provider_reference;
DROP TABLE IF EXISTS webhook_events;

DROP INDEX IF EXISTS idx_bank_accounts_currency;
DROP INDEX IF EXISTS idx_bank_accounts_user_id;
DROP TABLE IF EXISTS bank_accounts;

DROP INDEX IF EXISTS idx_idempotency_keys_transaction_id;
DROP TABLE IF EXISTS idempotency_keys;

DROP INDEX IF EXISTS idx_ledger_entries_created_at;
DROP INDEX IF EXISTS idx_ledger_entries_transaction_id;
DROP INDEX IF EXISTS idx_ledger_entries_wallet_id;
DROP TABLE IF EXISTS ledger_entries;

DROP INDEX IF EXISTS idx_transactions_created_at;
DROP INDEX IF EXISTS idx_transactions_status;
DROP INDEX IF EXISTS idx_transactions_to_wallet_id;
DROP INDEX IF EXISTS idx_transactions_from_wallet_id;
DROP INDEX IF EXISTS idx_transactions_trace_id;
DROP INDEX IF EXISTS idx_transactions_idempotency_key;
DROP TABLE IF EXISTS transactions;

DROP INDEX IF EXISTS idx_wallets_currency;
DROP INDEX IF EXISTS idx_wallets_user_id;
DROP TABLE IF EXISTS wallets;

DROP INDEX IF EXISTS idx_users_created_at;
DROP TABLE IF EXISTS users;
