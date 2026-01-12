package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/models"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/IfedayoAwe/payment-processing-service/services/mocks"
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPaymentService_GetTransactionByID(t *testing.T) {
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

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		genTx := gen.Transaction{
			ID:             "tx_123",
			IdempotencyKey: "key_123",
			FromWalletID:   sql.NullString{String: "wallet_1", Valid: true},
			ToWalletID:     sql.NullString{String: "wallet_2", Valid: true},
			Type:           "internal",
			Amount:         10000,
			Currency:       "USD",
			Status:         "completed",
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		mockQueries.On("GetTransactionByID", mock.Anything, "tx_123").Return(genTx, nil)

		result, err := ps.GetTransactionByID(context.Background(), "tx_123")

		require.NoError(t, err)
		assert.Equal(t, "tx_123", result.ID)
		assert.Equal(t, "completed", string(result.Status))
		mockQueries.AssertExpectations(t)
	})

	t.Run("transaction not found", func(t *testing.T) {
		mockQueries.On("GetTransactionByID", mock.Anything, "tx_notfound").Return(gen.Transaction{}, sql.ErrNoRows)

		result, err := ps.GetTransactionByID(context.Background(), "tx_notfound")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "transaction not found")
		mockQueries.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockQueries.On("GetTransactionByID", mock.Anything, "tx_error").Return(gen.Transaction{}, errors.New("db connection error"))

		result, err := ps.GetTransactionByID(context.Background(), "tx_error")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockQueries.AssertExpectations(t)
	})
}

func TestPaymentService_GetTransactionByIdempotencyKey(t *testing.T) {
	mockQueries := new(mocks.MockQuerier)
	processor := providers.NewProcessor()
	mockProvider := &mockCurrencyCloudProvider{}
	processor.RegisterPayoutProvider(mockProvider)
	processor.RegisterNameEnquiryProvider(mockProvider)
	processor.RegisterExchangeRateProvider(mockProvider)

	ps := &paymentService{
		queries:  mockQueries,
		provider: processor,
	}

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		genTx := gen.Transaction{
			ID:             "tx_123",
			IdempotencyKey: "key_123",
			FromWalletID:   sql.NullString{String: "wallet_1", Valid: true},
			Type:           "external",
			Amount:         5000,
			Currency:       "EUR",
			Status:         "initiated",
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_123").Return(genTx, nil)

		result, err := ps.GetTransactionByIdempotencyKey(context.Background(), "key_123")

		require.NoError(t, err)
		assert.Equal(t, "tx_123", result.ID)
		assert.Equal(t, "key_123", result.IdempotencyKey)
		mockQueries.AssertExpectations(t)
	})

	t.Run("not found returns sql.ErrNoRows", func(t *testing.T) {
		mockQueries.On("GetTransactionByIdempotencyKey", mock.Anything, "key_notfound").Return(gen.Transaction{}, sql.ErrNoRows)

		result, err := ps.GetTransactionByIdempotencyKey(context.Background(), "key_notfound")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, sql.ErrNoRows, err)
		mockQueries.AssertExpectations(t)
	})
}

func TestPaymentService_ConfirmTransaction_Validation(t *testing.T) {
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

	t.Run("transaction not found", func(t *testing.T) {
		mockQueries.On("GetTransactionByID", mock.Anything, "tx_notfound").Return(gen.Transaction{}, sql.ErrNoRows)

		result, err := ps.ConfirmTransaction(context.Background(), "tx_notfound", "user_1", "1234")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "transaction not found")
		mockQueries.AssertExpectations(t)
	})

	t.Run("transaction not in initiated status", func(t *testing.T) {
		now := time.Now()
		genTx := gen.Transaction{
			ID:           "tx_123",
			FromWalletID: sql.NullString{String: "wallet_1", Valid: true},
			Status:       "completed",
			CreatedAt:    now,
		}

		mockQueries.On("GetTransactionByID", mock.Anything, "tx_123").Return(genTx, nil)

		result, err := ps.ConfirmTransaction(context.Background(), "tx_123", "user_1", "1234")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "transaction is not in initiated status")
		mockQueries.AssertExpectations(t)
	})

	t.Run("transaction expired", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		processor.RegisterPayoutProvider(&mockCurrencyCloudProvider{})

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		oldTime := time.Now().Add(-11 * time.Minute)
		genTx := gen.Transaction{
			ID:           "tx_123",
			FromWalletID: sql.NullString{String: "wallet_1", Valid: true},
			Status:       "initiated",
			CreatedAt:    oldTime,
		}

		mockQueries.On("GetTransactionByID", mock.Anything, "tx_123").Return(genTx, nil)

		result, err := ps.ConfirmTransaction(context.Background(), "tx_123", "user_1", "1234")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "transaction has expired")
		mockQueries.AssertExpectations(t)
	})

	t.Run("transaction does not belong to user", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		processor.RegisterPayoutProvider(&mockCurrencyCloudProvider{})

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		now := time.Now()
		genTx := gen.Transaction{
			ID:           "tx_123",
			FromWalletID: sql.NullString{String: "wallet_1", Valid: true},
			Status:       "initiated",
			CreatedAt:    now,
		}

		wallet := &models.Wallet{
			ID:     "wallet_1",
			UserID: "user_2",
		}

		mockQueries.On("GetTransactionByID", mock.Anything, "tx_123").Return(genTx, nil)
		mockWallet.On("GetWalletByID", mock.Anything, "wallet_1").Return(wallet, nil)

		result, err := ps.ConfirmTransaction(context.Background(), "tx_123", "user_1", "1234")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "transaction does not belong to user")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		processor.RegisterPayoutProvider(&mockCurrencyCloudProvider{})

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		now := time.Now()
		genTx := gen.Transaction{
			ID:           "tx_123",
			FromWalletID: sql.NullString{String: "wallet_1", Valid: true},
			Status:       "initiated",
			CreatedAt:    now,
		}

		wallet := &models.Wallet{
			ID:     "wallet_1",
			UserID: "user_1",
		}

		mockQueries.On("GetTransactionByID", mock.Anything, "tx_123").Return(genTx, nil)
		mockWallet.On("GetWalletByID", mock.Anything, "wallet_1").Return(wallet, nil)
		mockQueries.On("GetUserByID", mock.Anything, "user_1").Return(gen.User{}, sql.ErrNoRows)

		result, err := ps.ConfirmTransaction(context.Background(), "tx_123", "user_1", "1234")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user not found")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
	})

	t.Run("PIN not set", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		processor.RegisterPayoutProvider(&mockCurrencyCloudProvider{})

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		now := time.Now()
		genTx := gen.Transaction{
			ID:           "tx_123",
			FromWalletID: sql.NullString{String: "wallet_1", Valid: true},
			Status:       "initiated",
			CreatedAt:    now,
		}

		wallet := &models.Wallet{
			ID:     "wallet_1",
			UserID: "user_1",
		}

		user := gen.User{
			UserID:  "user_1",
			PinHash: sql.NullString{Valid: false},
		}

		mockQueries.On("GetTransactionByID", mock.Anything, "tx_123").Return(genTx, nil)
		mockWallet.On("GetWalletByID", mock.Anything, "wallet_1").Return(wallet, nil)
		mockQueries.On("GetUserByID", mock.Anything, "user_1").Return(user, nil)

		result, err := ps.ConfirmTransaction(context.Background(), "tx_123", "user_1", "1234")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "PIN not set for user")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
	})

	t.Run("invalid PIN", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		mockWallet := new(mocks.MockWalletService)
		processor := providers.NewProcessor()
		processor.RegisterPayoutProvider(&mockCurrencyCloudProvider{})

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
			wallet:   mockWallet,
			ledger:   &ledgerService{queries: mockQueries},
		}

		now := time.Now()
		genTx := gen.Transaction{
			ID:           "tx_123",
			FromWalletID: sql.NullString{String: "wallet_1", Valid: true},
			Status:       "initiated",
			CreatedAt:    now,
		}

		wallet := &models.Wallet{
			ID:     "wallet_1",
			UserID: "user_1",
		}

		hashedPIN, _ := utils.HashPIN("5678")
		user := gen.User{
			UserID:  "user_1",
			PinHash: sql.NullString{String: hashedPIN, Valid: true},
		}

		mockQueries.On("GetTransactionByID", mock.Anything, "tx_123").Return(genTx, nil)
		mockWallet.On("GetWalletByID", mock.Anything, "wallet_1").Return(wallet, nil)
		mockQueries.On("GetUserByID", mock.Anything, "user_1").Return(user, nil)

		result, err := ps.ConfirmTransaction(context.Background(), "tx_123", "user_1", "1234")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid PIN")
		mockQueries.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
	})
}

func TestPaymentService_GetTransactionHistory(t *testing.T) {
	mockQueries := new(mocks.MockQuerier)
	processor := providers.NewProcessor()
	mockProvider := &mockCurrencyCloudProvider{}
	processor.RegisterPayoutProvider(mockProvider)
	processor.RegisterNameEnquiryProvider(mockProvider)
	processor.RegisterExchangeRateProvider(mockProvider)

	ps := &paymentService{
		queries:  mockQueries,
		provider: processor,
	}

	t.Run("success with limit", func(t *testing.T) {
		now := time.Now()
		transactions := []gen.Transaction{
			{
				ID:           "tx_1",
				FromWalletID: sql.NullString{String: "wallet_1", Valid: true},
				Amount:       10000,
				Currency:     "USD",
				Status:       "completed",
				CreatedAt:    now,
			},
			{
				ID:           "tx_2",
				FromWalletID: sql.NullString{String: "wallet_1", Valid: true},
				Amount:       5000,
				Currency:     "EUR",
				Status:       "pending",
				CreatedAt:    now.Add(-1 * time.Hour),
			},
		}

		mockQueries.On("ListTransactionsByUser", mock.Anything, mock.MatchedBy(func(arg gen.ListTransactionsByUserParams) bool {
			return arg.UserID == "user_1" && arg.Limit == 21
		})).Return(transactions, nil)

		result, err := ps.GetTransactionHistory(context.Background(), "user_1", "", 20)

		require.NoError(t, err)
		assert.Len(t, result.Transactions, 2)
		assert.Nil(t, result.NextCursor)
		mockQueries.AssertExpectations(t)
	})

	t.Run("with cursor", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		processor := providers.NewProcessor()
		processor.RegisterPayoutProvider(&mockCurrencyCloudProvider{})

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
		}

		now := time.Now()
		cursorTime := now.Add(-2 * time.Hour)
		transactions := []gen.Transaction{
			{
				ID:           "tx_3",
				FromWalletID: sql.NullString{String: "wallet_1", Valid: true},
				Amount:       3000,
				Currency:     "GBP",
				Status:       "completed",
				CreatedAt:    cursorTime,
			},
		}

		cursor := utils.EncodeCursor(cursorTime, "tx_prev")

		mockQueries.On("ListTransactionsByUser", mock.Anything, mock.MatchedBy(func(arg gen.ListTransactionsByUserParams) bool {
			return arg.UserID == "user_1"
		})).Return(transactions, nil)

		result, err := ps.GetTransactionHistory(context.Background(), "user_1", cursor, 20)

		require.NoError(t, err)
		assert.Len(t, result.Transactions, 1)
		mockQueries.AssertExpectations(t)
	})

	t.Run("limit validation", func(t *testing.T) {
		mockQueries := new(mocks.MockQuerier)
		processor := providers.NewProcessor()
		processor.RegisterPayoutProvider(&mockCurrencyCloudProvider{})

		ps := &paymentService{
			queries:  mockQueries,
			provider: processor,
		}

		mockQueries.On("ListTransactionsByUser", mock.Anything, mock.Anything).Return([]gen.Transaction{}, nil).Maybe()

		result, err := ps.GetTransactionHistory(context.Background(), "user_1", "", 0)
		require.NoError(t, err)
		assert.NotNil(t, result)

		result, err = ps.GetTransactionHistory(context.Background(), "user_1", "", 150)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}
