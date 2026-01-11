package services

import (
	"context"
	"testing"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/stretchr/testify/assert"
)

func TestLedgerService_CreateDebitEntry_Validation(t *testing.T) {
	ls := &ledgerService{queries: &gen.Queries{}}

	t.Run("positive amount should fail", func(t *testing.T) {
		err := ls.CreateDebitEntry(context.Background(), nil, "wallet_1", "tx_1", 1000, money.USD)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "debit amount must be negative")
	})

	t.Run("zero amount should fail", func(t *testing.T) {
		err := ls.CreateDebitEntry(context.Background(), nil, "wallet_1", "tx_1", 0, money.USD)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "debit amount must be negative")
	})

}

func TestLedgerService_CreateCreditEntry_Validation(t *testing.T) {
	ls := &ledgerService{queries: &gen.Queries{}}

	t.Run("negative amount should fail", func(t *testing.T) {
		err := ls.CreateCreditEntry(context.Background(), nil, "wallet_1", "tx_1", -1000, money.USD)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credit amount must be positive")
	})

	t.Run("zero amount should fail", func(t *testing.T) {
		err := ls.CreateCreditEntry(context.Background(), nil, "wallet_1", "tx_1", 0, money.USD)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credit amount must be positive")
	})

}
