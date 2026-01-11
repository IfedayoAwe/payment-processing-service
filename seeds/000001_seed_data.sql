-- Seed data for testing
-- Idempotent: Can be run multiple times without creating duplicates

SET timezone = 'UTC';

-- ============================================
-- USERS
-- ============================================
INSERT INTO users (user_id, name, pin_hash, created_at, updated_at)
VALUES 
    ('user_1', 'John Doe', '$2a$10$z1vJ04atupmQEzE6En2J5O7KtbhR3LgnbL7o2LizR9AhPE7cmA3wC', NOW(), NOW()),
    ('user_2', 'Jane Doe', '$2a$10$z1vJ04atupmQEzE6En2J5O7KtbhR3LgnbL7o2LizR9AhPE7cmA3wC', NOW(), NOW()),
    ('company_grey', 'Grey Company', NULL, NOW(), NOW())
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

-- Company bank accounts (Grey Company)
INSERT INTO bank_accounts (id, user_id, bank_name, bank_code, account_number, account_name, currency, provider, created_at, updated_at)
VALUES 
    ('bank_acc_company_usd', 'company_grey', 'Grey Bank', '044', '3000000001', 'Grey Company', 'USD', 'currencycloud', NOW(), NOW()),
    ('bank_acc_company_eur', 'company_grey', 'Grey Bank', '044', '3000000002', 'Grey Company', 'EUR', 'currencycloud', NOW(), NOW()),
    ('bank_acc_company_gbp', 'company_grey', 'Grey Bank', '044', '3000000003', 'Grey Company', 'GBP', 'dlocal', NOW(), NOW())
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

-- Company wallets (Grey Company): USD, EUR, GBP
-- Company should have more than both users combined (user_1: 1M USD, 850K EUR, 750K GBP; user_2: 0)
-- So company gets: 2M USD, 2M EUR, 2M GBP (more than combined)
INSERT INTO wallets (id, user_id, bank_account_id, currency, balance, created_at, updated_at)
VALUES 
    ('wallet_company_usd', 'company_grey', 'bank_acc_company_usd', 'USD', 0, NOW(), NOW()),
    ('wallet_company_eur', 'company_grey', 'bank_acc_company_eur', 'EUR', 0, NOW(), NOW()),
    ('wallet_company_gbp', 'company_grey', 'bank_acc_company_gbp', 'GBP', 0, NOW(), NOW())
ON CONFLICT (user_id, currency, bank_account_id) DO NOTHING;

-- ============================================
-- TRANSACTIONS (user_1 only)
-- ============================================
-- Initial funding transactions (self-transfers representing deposits)
-- These are placeholders for the ledger entries that credit the wallets
INSERT INTO transactions (id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency, status, provider_name, provider_reference, exchange_rate, created_at, updated_at)
VALUES 
    ('tx_user1_fund_usd', 'idemp_key_fund_usd', 'wallet_user1_usd', 'wallet_user1_usd', 'internal', 1000000, 'USD', 'completed', NULL, NULL, 1.00000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
    ('tx_user1_fund_eur', 'idemp_key_fund_eur', 'wallet_user1_eur', 'wallet_user1_eur', 'internal', 850000, 'EUR', 'completed', NULL, NULL, 1.00000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
    ('tx_user1_fund_gbp', 'idemp_key_fund_gbp', 'wallet_user1_gbp', 'wallet_user1_gbp', 'internal', 750000, 'GBP', 'completed', NULL, NULL, 1.00000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days')
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
-- Initial funding ledger entries (credits to give John Doe starting balances)
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, created_at)
VALUES 
    ('ledger_fund_usd', 'wallet_user1_usd', 'tx_user1_fund_usd', 1000000, 'USD', NOW() - INTERVAL '5 days'),
    ('ledger_fund_eur', 'wallet_user1_eur', 'tx_user1_fund_eur', 850000, 'EUR', NOW() - INTERVAL '5 days'),
    ('ledger_fund_gbp', 'wallet_user1_gbp', 'tx_user1_fund_gbp', 750000, 'GBP', NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

-- Transaction 1: USD -> EUR (internal)
-- Debit: -$58,823.53 from USD wallet (50000 EUR / 0.85 = 58823.53 USD)
-- Credit: +€50,000.00 to EUR wallet
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, created_at)
VALUES 
    ('ledger_1_debit', 'wallet_user1_usd', 'tx_user1_1', -58824, 'USD', NOW() - INTERVAL '2 days'),
    ('ledger_1_credit', 'wallet_user1_eur', 'tx_user1_1', 50000, 'EUR', NOW() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

-- Transaction 2: EUR -> GBP (internal)
-- Debit: -€33,970.59 from EUR wallet (30000 GBP / 0.88235294 = 33970.59 EUR)
-- Credit: +£30,000.00 to GBP wallet
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, created_at)
VALUES 
    ('ledger_2_debit', 'wallet_user1_eur', 'tx_user1_2', -33971, 'EUR', NOW() - INTERVAL '1 day'),
    ('ledger_2_credit', 'wallet_user1_gbp', 'tx_user1_2', 30000, 'GBP', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- Transaction 3: External USD transfer
-- Debit: -$100,000.00 from user USD wallet
-- Credit: +$100,000.00 to company USD wallet (double-entry)
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, created_at)
VALUES 
    ('ledger_3_debit', 'wallet_user1_usd', 'tx_user1_3', -100000, 'USD', NOW() - INTERVAL '12 hours'),
    ('ledger_3_credit', 'wallet_company_usd', 'tx_user1_3', 100000, 'USD', NOW() - INTERVAL '12 hours')
ON CONFLICT (id) DO NOTHING;

-- Transaction 4: External EUR transfer
-- Debit: -€50,000.00 from user EUR wallet
-- Credit: +€50,000.00 to company EUR wallet (double-entry)
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, created_at)
VALUES 
    ('ledger_4_debit', 'wallet_user1_eur', 'tx_user1_4', -50000, 'EUR', NOW() - INTERVAL '6 hours'),
    ('ledger_4_credit', 'wallet_company_eur', 'tx_user1_4', 50000, 'EUR', NOW() - INTERVAL '6 hours')
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- COMPANY FUNDING (initial funding for company wallets)
-- ============================================
-- Company funding transactions
INSERT INTO transactions (id, idempotency_key, from_wallet_id, to_wallet_id, type, amount, currency, status, provider_name, provider_reference, exchange_rate, created_at, updated_at)
VALUES 
    ('tx_company_fund_usd', 'idemp_key_company_fund_usd', 'wallet_company_usd', 'wallet_company_usd', 'internal', 2000000, 'USD', 'completed', NULL, NULL, 1.00000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
    ('tx_company_fund_eur', 'idemp_key_company_fund_eur', 'wallet_company_eur', 'wallet_company_eur', 'internal', 2000000, 'EUR', 'completed', NULL, NULL, 1.00000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
    ('tx_company_fund_gbp', 'idemp_key_company_fund_gbp', 'wallet_company_gbp', 'wallet_company_gbp', 'internal', 2000000, 'GBP', 'completed', NULL, NULL, 1.00000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

-- Company funding ledger entries
INSERT INTO ledger_entries (id, wallet_id, transaction_id, amount, currency, created_at)
VALUES 
    ('ledger_company_fund_usd', 'wallet_company_usd', 'tx_company_fund_usd', 2000000, 'USD', NOW() - INTERVAL '5 days'),
    ('ledger_company_fund_eur', 'wallet_company_eur', 'tx_company_fund_eur', 2000000, 'EUR', NOW() - INTERVAL '5 days'),
    ('ledger_company_fund_gbp', 'wallet_company_gbp', 'tx_company_fund_gbp', 2000000, 'GBP', NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

-- Company idempotency keys
INSERT INTO idempotency_keys (key, transaction_id, created_at)
VALUES 
    ('idemp_key_company_fund_usd', 'tx_company_fund_usd', NOW() - INTERVAL '5 days'),
    ('idemp_key_company_fund_eur', 'tx_company_fund_eur', NOW() - INTERVAL '5 days'),
    ('idemp_key_company_fund_gbp', 'tx_company_fund_gbp', NOW() - INTERVAL '5 days')
ON CONFLICT (key) DO NOTHING;

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
-- Recalculate balances based on ledger entries (source of truth)
-- This ensures wallet.balance matches the sum of ledger entries
UPDATE wallets 
SET balance = (
    SELECT COALESCE(SUM(amount), 0)
    FROM ledger_entries
    WHERE ledger_entries.wallet_id = wallets.id
),
updated_at = NOW()
WHERE id IN ('wallet_user1_usd', 'wallet_user1_eur', 'wallet_user1_gbp', 'wallet_user2_usd', 'wallet_user2_eur', 'wallet_user2_gbp', 'wallet_company_usd', 'wallet_company_eur', 'wallet_company_gbp');
