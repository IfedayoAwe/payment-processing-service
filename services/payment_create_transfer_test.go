package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/models"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/IfedayoAwe/payment-processing-service/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPaymentService_CreateInternalTransfer_Comprehensive(t *testing.T) {
	t.Run("idempotency key returns existing transaction", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		processor := providers.NewProcessor()
		mockProvider := &mockCurrencyCloudProvider{}
		processor.RegisterPayoutProvider(mockProvider)
		processor.RegisterNameEnquiryProvider(mockProvider)
		processor.RegisterExchangeRateProvider(mockProvider)

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   &walletService{queries: mockQueries},
			ledger:   &ledgerService{queries: mockQueries},
		}
		now := time.Now()
		existingTx := gen.Transaction{
			ID:             "tx_existing",
			IdempotencyKey: "key_123",
			FromWalletID:   sql.NullString{String: "wallet_1", Valid: true},
			Type:           "internal",
			Amount:         10000,
			Currency:       "USD",
			Status:         "completed",
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_123").Return(existingTx, nil)

		amount := money.NewMoney(10000, money.USD)
		result, err := ps.CreateInternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, amount, "key_123")

		require.NoError(t, err)
		assert.Equal(t, "tx_existing", result.ID)
		mockQueries.AssertExpectations(t)
	})

	t.Run("sender wallet not found", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		mockProvider := &mockCurrencyCloudProvider{}
		processor.RegisterPayoutProvider(mockProvider)
		processor.RegisterNameEnquiryProvider(mockProvider)
		processor.RegisterExchangeRateProvider(mockProvider)

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_new").Return(gen.Transaction{}, sql.ErrNoRows)
		mockWallet.On("GetWalletByUserAndCurrency", mock.Anything, "user_1", money.USD).Return(nil, sql.ErrNoRows)

		amount := money.NewMoney(10000, money.USD)
		result, err := ps.CreateInternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, amount, "key_new")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "sender wallet not found")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
	})

	t.Run("recipient account not found", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		mockProvider := &mockCurrencyCloudProvider{}
		processor.RegisterPayoutProvider(mockProvider)
		processor.RegisterNameEnquiryProvider(mockProvider)
		processor.RegisterExchangeRateProvider(mockProvider)

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		fromWallet := &models.Wallet{
			ID:       "wallet_1",
			UserID:   "user_1",
			Currency: "USD",
		}

		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_new").Return(gen.Transaction{}, sql.ErrNoRows)
		mockWallet.On("GetWalletByUserAndCurrency", mock.Anything, "user_1", money.USD).Return(fromWallet, nil)
		mockQueries.On("GetBankAccountByAccountAndBankCode", mock.Anything, mock.MatchedBy(func(arg gen.GetBankAccountByAccountAndBankCodeParams) bool {
			return arg.AccountNumber == "1234567890" && arg.BankCode == "044"
		})).Return(gen.BankAccount{}, sql.ErrNoRows)

		amount := money.NewMoney(10000, money.USD)
		result, err := ps.CreateInternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, amount, "key_new")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "recipient account not found")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
	})

	t.Run("currency mismatch", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		mockProvider := &mockCurrencyCloudProvider{}
		processor.RegisterPayoutProvider(mockProvider)
		processor.RegisterNameEnquiryProvider(mockProvider)
		processor.RegisterExchangeRateProvider(mockProvider)

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		fromWallet := &models.Wallet{
			ID:       "wallet_1",
			UserID:   "user_1",
			Currency: "USD",
		}

		bankAccount := gen.BankAccount{
			ID:            "acc_1",
			AccountNumber: "1234567890",
			BankCode:      "044",
			Currency:      "EUR",
		}

		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_new").Return(gen.Transaction{}, sql.ErrNoRows)
		mockWallet.On("GetWalletByUserAndCurrency", mock.Anything, "user_1", money.USD).Return(fromWallet, nil)
		mockQueries.On("GetBankAccountByAccountAndBankCode", mock.Anything, mock.Anything).Return(bankAccount, nil)

		amount := money.NewMoney(10000, money.USD)
		result, err := ps.CreateInternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, amount, "key_new")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "recipient account currency mismatch")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
	})

	t.Run("recipient wallet not found", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		mockProvider := &mockCurrencyCloudProvider{}
		processor.RegisterPayoutProvider(mockProvider)
		processor.RegisterNameEnquiryProvider(mockProvider)
		processor.RegisterExchangeRateProvider(mockProvider)

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		fromWallet := &models.Wallet{
			ID:       "wallet_1",
			UserID:   "user_1",
			Currency: "USD",
		}

		bankAccount := gen.BankAccount{
			ID:            "acc_1",
			AccountNumber: "1234567890",
			BankCode:      "044",
			Currency:      "USD",
		}

		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_new").Return(gen.Transaction{}, sql.ErrNoRows)
		mockWallet.On("GetWalletByUserAndCurrency", mock.Anything, "user_1", money.USD).Return(fromWallet, nil)
		mockQueries.On("GetBankAccountByAccountAndBankCode", mock.Anything, mock.Anything).Return(bankAccount, nil)
		mockWallet.On("GetWalletByBankAccount", mock.Anything, "acc_1").Return(nil, sql.ErrNoRows)

		amount := money.NewMoney(10000, money.USD)
		result, err := ps.CreateInternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, amount, "key_new")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "recipient wallet not found for bank account")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
	})

	t.Run("cannot transfer to same wallet", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		mockProvider := &mockCurrencyCloudProvider{}
		processor.RegisterPayoutProvider(mockProvider)
		processor.RegisterNameEnquiryProvider(mockProvider)
		processor.RegisterExchangeRateProvider(mockProvider)

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		fromWallet := &models.Wallet{
			ID:       "wallet_1",
			UserID:   "user_1",
			Currency: "USD",
		}

		toWallet := &models.Wallet{
			ID:       "wallet_1",
			UserID:   "user_1",
			Currency: "USD",
		}

		bankAccount := gen.BankAccount{
			ID:            "acc_1",
			AccountNumber: "1234567890",
			BankCode:      "044",
			Currency:      "USD",
		}

		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_new").Return(gen.Transaction{}, sql.ErrNoRows)
		mockWallet.On("GetWalletByUserAndCurrency", mock.Anything, "user_1", money.USD).Return(fromWallet, nil)
		mockQueries.On("GetBankAccountByAccountAndBankCode", mock.Anything, mock.Anything).Return(bankAccount, nil)
		mockWallet.On("GetWalletByBankAccount", mock.Anything, "acc_1").Return(toWallet, nil)

		amount := money.NewMoney(10000, money.USD)
		result, err := ps.CreateInternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, amount, "key_new")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cannot transfer to same wallet")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
	})

	t.Run("idempotency check database error", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		processor := providers.NewProcessor()
		mockProvider := &mockCurrencyCloudProvider{}
		processor.RegisterPayoutProvider(mockProvider)
		processor.RegisterNameEnquiryProvider(mockProvider)
		processor.RegisterExchangeRateProvider(mockProvider)

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   &walletService{queries: mockQueries},
			ledger:   &ledgerService{queries: mockQueries},
		}

		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_error").Return(gen.Transaction{}, errors.New("db error"))

		amount := money.NewMoney(10000, money.USD)
		result, err := ps.CreateInternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, amount, "key_error")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "check idempotency")
		mockQueries.AssertExpectations(t)
	})
}

func TestPaymentService_CreateExternalTransfer_Comprehensive(t *testing.T) {
	t.Run("insufficient funds", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		mockLedger := new(mocks.MockLedgerService)
		processor := providers.NewProcessor()
		mockProvider := &mockCurrencyCloudProvider{}
		processor.RegisterPayoutProvider(mockProvider)
		processor.RegisterNameEnquiryProvider(mockProvider)
		processor.RegisterExchangeRateProvider(mockProvider)

		externalTransferSvc := &externalTransferService{
			queries:  mockQueries,
			wallet:   mockWallet,
			ledger:   mockLedger,
			provider: processor,
		}

		ps := &paymentService{
			queries:          mockQueries,
			provider:         processor,
			wallet:           mockWallet,
			ledger:           mockLedger,
			externalTransfer: externalTransferSvc,
		}

		fromWallet := &models.Wallet{
			ID:       "wallet_1",
			UserID:   "user_1",
			Currency: "USD",
			Balance:  1000,
		}

		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_1").Return(gen.Transaction{}, sql.ErrNoRows)
		mockWallet.On("GetWalletByUserAndCurrency", mock.Anything, "user_1", money.USD).Return(fromWallet, nil)

		amount := money.NewMoney(10000, money.USD)
		result, err := ps.CreateExternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, amount, "key_1")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "insufficient funds")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
		mockLedger.AssertExpectations(t)
	})
}
