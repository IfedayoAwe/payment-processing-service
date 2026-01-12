-- Seed data for testing
-- Idempotent: Can be run multiple times without creating duplicates
--
-- Balance Validation:
-- - All transactions have balanced ledger entries (debits = credits per currency)
-- - External transactions: debit user wallet, credit external system (balances external system)
-- - Internal transactions: debit sender wallet, credit receiver wallet (same currency amounts)
-- - External system balances are cumulative per currency (negative = we owe external system)
-- - Wallet balances are calculated from ledger entries at the end

SET timezone = 'UTC';

-- ============================================
-- USERS
-- ============================================
INSERT INTO users (user_id, name, pin_hash, created_at, updated_at)
VALUES 
    ('user_1', 'John Doe', '$2a$10$z1vJ04atupmQEzE6En2J5O7KtbhR3LgnbL7o2LizR9AhPE7cmA3wC', NOW(), NOW()),
    ('user_2', 'Jane Doe', '$2a$10$z1vJ04atupmQEzE6En2J5O7KtbhR3LgnbL7o2LizR9AhPE7cmA3wC', NOW(), NOW())
ON CONFLICT (user_id) DO NOTHING;

-- ============================================
-- BANK ACCOUNTS
-- ============================================
-- CurrencyCloud accounts for user_1 (John Doe)
INSERT INTO bank_accounts (id, user_id, bank_name, bank_code, account_number, account_name, currency, provider, created_at, updated_at)
VALUES 
    ('bank_acc_user1_usd_cc', 'user_1', 'Test Bank', '044', '1000000001', 'John Doe', 'USD', 'currencycloud', NOW(), NOW()),
    ('bank_acc_user1_eur_cc', 'user_1', 'Test Bank', '044', '1000000002', 'John Doe', 'EUR', 'currencycloud', NOW(), NOW())
ON CONFLICT (account_number, bank_code) DO NOTHING;

-- dLocal accounts for user_1 (John Doe)
INSERT INTO bank_accounts (id, user_id, bank_name, bank_code, account_number, account_name, currency, provider, created_at, updated_at)
VALUES 
    ('bank_acc_user1_usd_dl', 'user_1', 'Test Bank', '044', '1000000003', 'John Doe', 'USD', 'dlocal', NOW(), NOW()),
    ('bank_acc_user1_eur_dl', 'user_1', 'Test Bank', '044', '1000000004', 'John Doe', 'EUR', 'dlocal', NOW(), NOW()),
    ('bank_acc_user1_gbp_dl', 'user_1', 'Test Bank', '044', '1000000005', 'John Doe', 'GBP', 'dlocal', NOW(), NOW())
ON CONFLICT (account_number, bank_code) DO NOTHING;

-- CurrencyCloud accounts for user_2 (Jane Doe)
INSERT INTO bank_accounts (id, user_id, bank_name, bank_code, account_number, account_name, currency, provider, created_at, updated_at)
VALUES 
    ('bank_acc_user2_usd_cc', 'user_2', 'Test Bank', '044', '2000000001', 'Jane Doe', 'USD', 'currencycloud', NOW(), NOW()),
    ('bank_acc_user2_eur_cc', 'user_2', 'Test Bank', '044', '2000000002', 'Jane Doe', 'EUR', 'currencycloud', NOW(), NOW())
ON CONFLICT (account_number, bank_code) DO NOTHING;

-- dLocal accounts for user_2 (Jane Doe)
INSERT INTO bank_accounts (id, user_id, bank_name, bank_code, account_number, account_name, currency, provider, created_at, updated_at)
VALUES 
    ('bank_acc_user2_usd_dl', 'user_2', 'Test Bank', '044', '2000000003', 'Jane Doe', 'USD', 'dlocal', NOW(), NOW()),
    ('bank_acc_user2_eur_dl', 'user_2', 'Test Bank', '044', '2000000004', 'Jane Doe', 'EUR', 'dlocal', NOW(), NOW()),
    ('bank_acc_user2_gbp_dl', 'user_2', 'Test Bank', '044', '2000000005', 'Jane Doe', 'GBP', 'dlocal', NOW(), NOW())
ON CONFLICT (account_number, bank_code) DO NOTHING;

-- ============================================
-- WALLETS (initialized with 0 balance, will be updated from ledger)
-- ============================================
-- user_1 (John Doe): USD and EUR with CurrencyCloud, GBP with dLocal
INSERT INTO wallets (id, user_id, bank_account_id, currency, balance, created_at, updated_at)
VALUES 
    ('wallet_user1_usd', 'user_1', 'bank_acc_user1_usd_cc', 'USD', 0, NOW(), NOW()),
    ('wallet_user1_eur', 'user_1', 'bank_acc_user1_eur_cc', 'EUR', 0, NOW(), NOW()),
    ('wallet_user1_gbp', 'user_1', 'bank_acc_user1_gbp_dl', 'GBP', 0, NOW(), NOW())
ON CONFLICT (user_id, currency, bank_account_id) DO NOTHING;

-- user_2 (Jane Doe): USD and EUR with CurrencyCloud, GBP with dLocal
INSERT INTO wallets (id, user_id, bank_account_id, currency, balance, created_at, updated_at)
VALUES 
    ('wallet_user2_usd', 'user_2', 'bank_acc_user2_usd_cc', 'USD', 0, NOW(), NOW()),
    ('wallet_user2_eur', 'user_2', 'bank_acc_user2_eur_cc', 'EUR', 0, NOW(), NOW()),
    ('wallet_user2_gbp', 'user_2', 'bank_acc_user2_gbp_dl', 'GBP', 0, NOW(), NOW())
ON CONFLICT (user_id, currency, bank_account_id) DO NOTHING;

-- ============================================
-- TRANSACTIONS (user_1 only)
-- ============================================
-- Initial funding transactions (external deposits from external system)
INSERT INTO transactions (id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency, status, provider_name, provider_reference, exchange_rate, created_at, updated_at)
VALUES 
    ('tx_user1_fund_usd', 'idemp_key_fund_usd', NULL, 'wallet_user1_usd', 'external', 1000000, 'USD', 'completed', 'currencycloud', 'DEPOSIT-USD-001', 1.00000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
    ('tx_user1_fund_eur', 'idemp_key_fund_eur', NULL, 'wallet_user1_eur', 'external', 850000, 'EUR', 'completed', 'currencycloud', 'DEPOSIT-EUR-001', 1.00000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
    ('tx_user1_fund_gbp', 'idemp_key_fund_gbp', NULL, 'wallet_user1_gbp', 'external', 750000, 'GBP', 'completed', 'dlocal', 'DEPOSIT-GBP-001', 1.00000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

-- Transaction 1: Internal transfer USD -> EUR (completed)
INSERT INTO transactions (id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency, status, provider_name, provider_reference, exchange_rate, created_at, updated_at)
VALUES 
    ('tx_user1_1', 'idemp_key_1', 'wallet_user1_usd', 'wallet_user1_eur', 'internal', 50000, 'EUR', 'completed', NULL, NULL, 0.85000000, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

-- Transaction 2: Internal transfer EUR -> GBP (completed)
INSERT INTO transactions (id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency, status, provider_name, provider_reference, exchange_rate, created_at, updated_at)
VALUES 
    ('tx_user1_2', 'idemp_key_2', 'wallet_user1_eur', 'wallet_user1_gbp', 'internal', 30000, 'GBP', 'completed', NULL, NULL, 0.88235294, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- Transaction 3: External transfer USD (completed)
INSERT INTO transactions (id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency, status, provider_name, provider_reference, exchange_rate, created_at, updated_at)
VALUES 
    ('tx_user1_3', 'idemp_key_3', 'wallet_user1_usd', NULL, 'external', 100000, 'USD', 'completed', 'currencycloud', 'TXN-tx_user1-1234567890', 1.00000000, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours')
ON CONFLICT (id) DO NOTHING;

-- Transaction 4: External transfer EUR (completed)
INSERT INTO transactions (id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency, status, provider_name, provider_reference, exchange_rate, created_at, updated_at)
VALUES 
    ('tx_user1_4', 'idemp_key_4', 'wallet_user1_eur', NULL, 'external', 50000, 'EUR', 'completed', 'currencycloud', 'TXN-tx_user1-1234567891', 1.00000000, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours')
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- LEDGER ENTRIES (user_1 only)
-- ============================================
-- Initial funding ledger entries (external deposits: debit external system, credit user wallets)
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, account_type, balance_before, balance_after, created_at)
VALUES 
    ('ledger_fund_usd_debit', NULL, 'tx_user1_fund_usd', -1000000, 'USD', 'external_wallet', 0, -1000000, NOW() - INTERVAL '5 days'),
    ('ledger_fund_usd_credit', 'wallet_user1_usd', 'tx_user1_fund_usd', 1000000, 'USD', 'user_wallet', 0, 1000000, NOW() - INTERVAL '5 days'),
    ('ledger_fund_eur_debit', NULL, 'tx_user1_fund_eur', -850000, 'EUR', 'external_wallet', 0, -850000, NOW() - INTERVAL '5 days'),
    ('ledger_fund_eur_credit', 'wallet_user1_eur', 'tx_user1_fund_eur', 850000, 'EUR', 'user_wallet', 0, 850000, NOW() - INTERVAL '5 days'),
    ('ledger_fund_gbp_debit', NULL, 'tx_user1_fund_gbp', -750000, 'GBP', 'external_wallet', 0, -750000, NOW() - INTERVAL '5 days'),
    ('ledger_fund_gbp_credit', 'wallet_user1_gbp', 'tx_user1_fund_gbp', 750000, 'GBP', 'user_wallet', 0, 750000, NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

-- Transaction 1: USD -> EUR (internal)
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, account_type, balance_before, balance_after, created_at)
VALUES 
    ('ledger_1_debit', 'wallet_user1_usd', 'tx_user1_1', -58824, 'USD', 'user_wallet', 1000000, 941176, NOW() - INTERVAL '2 days'),
    ('ledger_1_credit', 'wallet_user1_eur', 'tx_user1_1', 50000, 'EUR', 'user_wallet', 850000, 900000, NOW() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

-- Transaction 2: EUR -> GBP (internal)
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, account_type, balance_before, balance_after, created_at)
VALUES 
    ('ledger_2_debit', 'wallet_user1_eur', 'tx_user1_2', -33971, 'EUR', 'user_wallet', 900000, 866029, NOW() - INTERVAL '1 day'),
    ('ledger_2_credit', 'wallet_user1_gbp', 'tx_user1_2', 30000, 'GBP', 'user_wallet', 750000, 780000, NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- Transaction 3: External USD transfer (debit user wallet, credit external system)
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, account_type, balance_before, balance_after, created_at)
VALUES 
    ('ledger_3_debit', 'wallet_user1_usd', 'tx_user1_3', -100000, 'USD', 'user_wallet', 941176, 841176, NOW() - INTERVAL '12 hours'),
    ('ledger_3_credit', NULL, 'tx_user1_3', 100000, 'USD', 'external_wallet', -1000000, -900000, NOW() - INTERVAL '12 hours')
ON CONFLICT (id) DO NOTHING;

-- Transaction 4: External EUR transfer (debit user wallet, credit external system)
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, account_type, balance_before, balance_after, created_at)
VALUES 
    ('ledger_4_debit', 'wallet_user1_eur', 'tx_user1_4', -50000, 'EUR', 'user_wallet', 866029, 816029, NOW() - INTERVAL '6 hours'),
    ('ledger_4_credit', NULL, 'tx_user1_4', 50000, 'EUR', 'external_wallet', -850000, -800000, NOW() - INTERVAL '6 hours')
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- IDEMPOTENCY KEYS (user_1 only)
-- ============================================
INSERT INTO idempotency_keys (key, transaction_id, created_at)
VALUES 
    ('idemp_key_fund_usd', 'tx_user1_fund_usd', NOW() - INTERVAL '5 days'),
    ('idemp_key_fund_eur', 'tx_user1_fund_eur', NOW() - INTERVAL '5 days'),
    ('idemp_key_fund_gbp', 'tx_user1_fund_gbp', NOW() - INTERVAL '5 days'),
    ('idemp_key_1', 'tx_user1_1', NOW() - INTERVAL '2 days'),
    ('idemp_key_2', 'tx_user1_2', NOW() - INTERVAL '1 day'),
    ('idemp_key_3', 'tx_user1_3', NOW() - INTERVAL '12 hours'),
    ('idemp_key_4', 'tx_user1_4', NOW() - INTERVAL '6 hours')
ON CONFLICT (key) DO NOTHING;

-- ============================================
-- UPDATE WALLET BALANCES (all wallets)
-- ============================================
UPDATE wallets 
SET balance = (
    SELECT COALESCE(SUM(amount), 0)
    FROM ledger_entries
    WHERE ledger_entries.wallet_id = wallets.id AND ledger_entries.account_type = 'user_wallet'
),
updated_at = NOW()
WHERE id IN ('wallet_user1_usd', 'wallet_user1_eur', 'wallet_user1_gbp', 'wallet_user2_usd', 'wallet_user2_eur', 'wallet_user2_gbp');
