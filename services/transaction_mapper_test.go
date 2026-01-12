package services

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapTransaction(t *testing.T) {
	now := time.Now()

	t.Run("complete transaction with all fields", func(t *testing.T) {
		exchangeRate := "1.25"
		genTx := gen.Transaction{
			ID:                "tx_123",
			IdempotencyKey:    "key_123",
			FromWalletID:      sql.NullString{String: "wallet_from", Valid: true},
			ToWalletID:        sql.NullString{String: "wallet_to", Valid: true},
			Type:              "internal",
			Amount:            10000,
			Currency:          "USD",
			Status:            "completed",
			ProviderName:      sql.NullString{String: "currencycloud", Valid: true},
			ProviderReference: sql.NullString{String: "ref_123", Valid: true},
			ExchangeRate:      sql.NullString{String: exchangeRate, Valid: true},
			FailureReason:     sql.NullString{Valid: false},
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		result := mapTransaction(genTx)

		require.NotNil(t, result)
		assert.Equal(t, "tx_123", result.ID)
		assert.Equal(t, "key_123", result.IdempotencyKey)
		assert.Equal(t, "wallet_from", *result.FromWalletID)
		assert.Equal(t, "wallet_to", *result.ToWalletID)
		assert.Equal(t, "internal", string(result.Type))
		assert.Equal(t, int64(10000), result.Amount)
		assert.Equal(t, "USD", result.Currency)
		assert.Equal(t, "completed", string(result.Status))
		assert.Equal(t, "currencycloud", *result.ProviderName)
		assert.Equal(t, "ref_123", *result.ProviderReference)
		require.NotNil(t, result.ExchangeRate)
		assert.Equal(t, 1.25, *result.ExchangeRate)
		assert.Nil(t, result.FailureReason)
	})

	t.Run("transaction with null optional fields", func(t *testing.T) {
		genTx := gen.Transaction{
			ID:                "tx_456",
			IdempotencyKey:    "key_456",
			FromWalletID:      sql.NullString{String: "wallet_from", Valid: true},
			ToWalletID:        sql.NullString{Valid: false},
			Type:              "external",
			Amount:            5000,
			Currency:          "EUR",
			Status:            "pending",
			ProviderName:      sql.NullString{Valid: false},
			ProviderReference: sql.NullString{Valid: false},
			ExchangeRate:      sql.NullString{Valid: false},
			FailureReason:     sql.NullString{String: "insufficient funds", Valid: true},
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		result := mapTransaction(genTx)

		require.NotNil(t, result)
		assert.Equal(t, "tx_456", result.ID)
		assert.Nil(t, result.ToWalletID)
		assert.Nil(t, result.ProviderName)
		assert.Nil(t, result.ProviderReference)
		assert.Nil(t, result.ExchangeRate)
		assert.Equal(t, "insufficient funds", *result.FailureReason)
	})

	t.Run("invalid exchange rate string", func(t *testing.T) {
		genTx := gen.Transaction{
			ID:             "tx_789",
			IdempotencyKey: "key_789",
			FromWalletID:   sql.NullString{String: "wallet_from", Valid: true},
			ToWalletID:     sql.NullString{Valid: false},
			Type:           "internal",
			Amount:         1000,
			Currency:       "GBP",
			Status:         "completed",
			ExchangeRate:   sql.NullString{String: "invalid_rate", Valid: true},
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		result := mapTransaction(genTx)

		require.NotNil(t, result)
		assert.Nil(t, result.ExchangeRate)
	})

	t.Run("zero exchange rate", func(t *testing.T) {
		genTx := gen.Transaction{
			ID:             "tx_zero",
			IdempotencyKey: "key_zero",
			FromWalletID:   sql.NullString{String: "wallet_from", Valid: true},
			ToWalletID:     sql.NullString{Valid: false},
			Type:           "internal",
			Amount:         1000,
			Currency:       "USD",
			Status:         "completed",
			ExchangeRate:   sql.NullString{String: "0", Valid: true},
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		result := mapTransaction(genTx)

		require.NotNil(t, result)
		assert.Nil(t, result.ExchangeRate)
	})

	t.Run("negative exchange rate", func(t *testing.T) {
		genTx := gen.Transaction{
			ID:             "tx_neg",
			IdempotencyKey: "key_neg",
			FromWalletID:   sql.NullString{String: "wallet_from", Valid: true},
			ToWalletID:     sql.NullString{Valid: false},
			Type:           "internal",
			Amount:         1000,
			Currency:       "USD",
			Status:         "completed",
			ExchangeRate:   sql.NullString{String: "-1.5", Valid: true},
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		result := mapTransaction(genTx)

		require.NotNil(t, result)
		assert.Nil(t, result.ExchangeRate)
	})

	t.Run("very precise exchange rate", func(t *testing.T) {
		rate := strconv.FormatFloat(1.23456789, 'f', 8, 64)
		genTx := gen.Transaction{
			ID:             "tx_precise",
			IdempotencyKey: "key_precise",
			FromWalletID:   sql.NullString{String: "wallet_from", Valid: true},
			ToWalletID:     sql.NullString{Valid: false},
			Type:           "internal",
			Amount:         1000,
			Currency:       "USD",
			Status:         "completed",
			ExchangeRate:   sql.NullString{String: rate, Valid: true},
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		result := mapTransaction(genTx)

		require.NotNil(t, result)
		require.NotNil(t, result.ExchangeRate)
		assert.InDelta(t, 1.23456789, *result.ExchangeRate, 0.00000001)
	})
}
