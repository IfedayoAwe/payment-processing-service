package services

import (
	"context"
	"testing"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/stretchr/testify/assert"
)

func TestPaymentService_GetExchangeRate(t *testing.T) {
	processor := providers.NewProcessor()
	mockProvider := &mockCurrencyCloudProvider{}
	processor.RegisterPayoutProvider(mockProvider)
	processor.RegisterNameEnquiryProvider(mockProvider)
	processor.RegisterExchangeRateProvider(mockProvider)

	ps := &paymentService{
		queries:  &gen.Queries{},
		provider: processor,
	}

	t.Run("success", func(t *testing.T) {
		rate, err := ps.GetExchangeRate(context.Background(), money.USD, money.EUR)
		assert.NoError(t, err)
		assert.Equal(t, 1.2, rate)
	})
}

func TestPaymentService_CreateInternalTransfer_Validation(t *testing.T) {
	processor := providers.NewProcessor()
	mockProvider := &mockCurrencyCloudProvider{}
	processor.RegisterPayoutProvider(mockProvider)
	processor.RegisterNameEnquiryProvider(mockProvider)
	processor.RegisterExchangeRateProvider(mockProvider)

	ps := &paymentService{
		queries:  &gen.Queries{},
		provider: processor,
	}

	t.Run("zero amount should fail", func(t *testing.T) {
		zeroAmount := money.NewMoney(0, money.USD)
		_, err := ps.CreateInternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, zeroAmount, "key_1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("negative amount should fail", func(t *testing.T) {
		negativeAmount := money.NewMoney(-100, money.USD)
		_, err := ps.CreateInternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, negativeAmount, "key_1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("different currencies", func(t *testing.T) {
		rate, err := ps.GetExchangeRate(context.Background(), money.EUR, money.GBP)
		assert.NoError(t, err)
		assert.Greater(t, rate, 0.0)
	})
}
