package mocks

import (
	"context"
	"database/sql"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/stretchr/testify/mock"
)

type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CreateIdempotencyKey(ctx context.Context, arg gen.CreateIdempotencyKeyParams) (gen.IdempotencyKey, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(gen.IdempotencyKey), args.Error(1)
}

func (m *MockQuerier) CreateLedgerEntry(ctx context.Context, arg gen.CreateLedgerEntryParams) (gen.LedgerEntry, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(gen.LedgerEntry), args.Error(1)
}

func (m *MockQuerier) CreateTransaction(ctx context.Context, arg gen.CreateTransactionParams) (gen.Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(gen.Transaction), args.Error(1)
}

func (m *MockQuerier) CreateWallet(ctx context.Context, arg gen.CreateWalletParams) (gen.Wallet, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(gen.Wallet), args.Error(1)
}

func (m *MockQuerier) CreateWebhookEvent(ctx context.Context, arg gen.CreateWebhookEventParams) (gen.WebhookEvent, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(gen.WebhookEvent), args.Error(1)
}

func (m *MockQuerier) GetBankAccountByAccountAndBankCode(ctx context.Context, arg gen.GetBankAccountByAccountAndBankCodeParams) (gen.BankAccount, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return gen.BankAccount{}, args.Error(1)
	}
	return args.Get(0).(gen.BankAccount), args.Error(1)
}

func (m *MockQuerier) GetBankAccountByID(ctx context.Context, id string) (gen.BankAccount, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return gen.BankAccount{}, args.Error(1)
	}
	return args.Get(0).(gen.BankAccount), args.Error(1)
}

func (m *MockQuerier) GetTransactionByID(ctx context.Context, id string) (gen.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return gen.Transaction{}, args.Error(1)
	}
	return args.Get(0).(gen.Transaction), args.Error(1)
}

func (m *MockQuerier) GetTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (gen.Transaction, error) {
	args := m.Called(ctx, idempotencyKey)
	if args.Get(0) == nil {
		return gen.Transaction{}, args.Error(1)
	}
	return args.Get(0).(gen.Transaction), args.Error(1)
}

func (m *MockQuerier) GetUserByID(ctx context.Context, userID string) (gen.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return gen.User{}, args.Error(1)
	}
	return args.Get(0).(gen.User), args.Error(1)
}

func (m *MockQuerier) GetUserWalletsWithBankAccounts(ctx context.Context, userID string) ([]gen.GetUserWalletsWithBankAccountsRow, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]gen.GetUserWalletsWithBankAccountsRow), args.Error(1)
}

func (m *MockQuerier) GetWalletBalance(ctx context.Context, arg gen.GetWalletBalanceParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) GetWalletByBankAccount(ctx context.Context, bankAccountID sql.NullString) (gen.Wallet, error) {
	args := m.Called(ctx, bankAccountID)
	if args.Get(0) == nil {
		return gen.Wallet{}, args.Error(1)
	}
	return args.Get(0).(gen.Wallet), args.Error(1)
}

func (m *MockQuerier) GetWalletByID(ctx context.Context, id string) (gen.Wallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return gen.Wallet{}, args.Error(1)
	}
	return args.Get(0).(gen.Wallet), args.Error(1)
}

func (m *MockQuerier) GetWalletByIDForUpdate(ctx context.Context, id string) (gen.Wallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return gen.Wallet{}, args.Error(1)
	}
	return args.Get(0).(gen.Wallet), args.Error(1)
}

func (m *MockQuerier) GetWalletByUserAndCurrency(ctx context.Context, arg gen.GetWalletByUserAndCurrencyParams) (gen.Wallet, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return gen.Wallet{}, args.Error(1)
	}
	return args.Get(0).(gen.Wallet), args.Error(1)
}

func (m *MockQuerier) GetWalletByUserAndCurrencyForUpdate(ctx context.Context, arg gen.GetWalletByUserAndCurrencyForUpdateParams) (gen.Wallet, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return gen.Wallet{}, args.Error(1)
	}
	return args.Get(0).(gen.Wallet), args.Error(1)
}

func (m *MockQuerier) ListTransactionsByUser(ctx context.Context, arg gen.ListTransactionsByUserParams) ([]gen.Transaction, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]gen.Transaction), args.Error(1)
}

func (m *MockQuerier) UpdateTransactionFailure(ctx context.Context, arg gen.UpdateTransactionFailureParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) UpdateTransactionStatus(ctx context.Context, arg gen.UpdateTransactionStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) UpdateTransactionWithProvider(ctx context.Context, arg gen.UpdateTransactionWithProviderParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) UpdateWalletBalance(ctx context.Context, arg gen.UpdateWalletBalanceParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) WithTx(tx *sql.Tx) gen.Querier {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(gen.Querier)
}
