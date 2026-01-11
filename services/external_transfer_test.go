package services

import (
	"context"
	"testing"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/stretchr/testify/assert"
)

func TestExternalTransferService_CreateExternalTransfer_Validation(t *testing.T) {
	processor := providers.NewProcessor()
	mockProvider := &mockCurrencyCloudProvider{}
	processor.RegisterPayoutProvider(mockProvider)
	processor.RegisterNameEnquiryProvider(mockProvider)
	processor.RegisterExchangeRateProvider(mockProvider)

	ets := &externalTransferService{
		queries:  &gen.Queries{},
		provider: processor,
	}

	t.Run("zero amount should fail", func(t *testing.T) {
		zeroAmount := money.NewMoney(0, money.USD)
		_, err := ets.CreateExternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, zeroAmount, 1.0, "key_1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("negative amount should fail", func(t *testing.T) {
		negativeAmount := money.NewMoney(-100, money.USD)
		_, err := ets.CreateExternalTransfer(context.Background(), "user_1", "1234567890", "044", money.USD, negativeAmount, 1.0, "key_1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})
}
