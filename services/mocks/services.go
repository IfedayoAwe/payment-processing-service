package mocks

import (
	"context"
	"database/sql"

	"github.com/IfedayoAwe/payment-processing-service/models"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/stretchr/testify/mock"
)

type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) GetWalletByID(ctx context.Context, walletID string) (*models.Wallet, error) {
	args := m.Called(ctx, walletID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletService) GetWalletByUserAndCurrency(ctx context.Context, userID string, currency money.Currency) (*models.Wallet, error) {
	args := m.Called(ctx, userID, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletService) GetWalletByBankAccount(ctx context.Context, bankAccountID string) (*models.Wallet, error) {
	args := m.Called(ctx, bankAccountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletService) GetUserWallets(ctx context.Context, userID string) ([]*models.WalletWithBankAccount, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.WalletWithBankAccount), args.Error(1)
}

func (m *MockWalletService) LockWalletForUpdate(ctx context.Context, tx *sql.Tx, walletID string) (*models.Wallet, error) {
	args := m.Called(ctx, tx, walletID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletService) LockWalletByUserAndCurrency(ctx context.Context, tx *sql.Tx, userID string, currency money.Currency) (*models.Wallet, error) {
	args := m.Called(ctx, tx, userID, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

type MockLedgerService struct {
	mock.Mock
}

func (m *MockLedgerService) CreateDebitEntry(ctx context.Context, tx *sql.Tx, walletID string, transactionID string, amount int64, currency money.Currency) error {
	args := m.Called(ctx, tx, walletID, transactionID, amount, currency)
	return args.Error(0)
}

func (m *MockLedgerService) CreateCreditEntry(ctx context.Context, tx *sql.Tx, walletID string, transactionID string, amount int64, currency money.Currency) error {
	args := m.Called(ctx, tx, walletID, transactionID, amount, currency)
	return args.Error(0)
}

func (m *MockLedgerService) GetWalletBalance(ctx context.Context, tx *sql.Tx, walletID string, currency money.Currency) (int64, error) {
	args := m.Called(ctx, tx, walletID, currency)
	return args.Get(0).(int64), args.Error(1)
}
