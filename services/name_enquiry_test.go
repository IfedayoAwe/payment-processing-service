package services

import (
	"context"
	"testing"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/stretchr/testify/assert"
)

func TestNameEnquiryService_EnquireAccountName_Validation(t *testing.T) {
	processor := providers.NewProcessor()
	mockProvider := &mockCurrencyCloudProvider{}
	processor.RegisterPayoutProvider(mockProvider)
	processor.RegisterNameEnquiryProvider(mockProvider)
	processor.RegisterExchangeRateProvider(mockProvider)

	nes := &nameEnquiryService{
		queries:  &gen.Queries{},
		provider: processor,
	}

	t.Run("missing account number", func(t *testing.T) {
		result, err := nes.EnquireAccountName(context.Background(), "", "044")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "account_number is required")
	})

	t.Run("missing bank code", func(t *testing.T) {
		result, err := nes.EnquireAccountName(context.Background(), "1234567890", "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "bank_code is required")
	})
}

type mockCurrencyCloudProvider struct{}

func (m *mockCurrencyCloudProvider) NameEnquiry(ctx context.Context, req providers.NameEnquiryRequest) (*providers.NameEnquiryResponse, error) {
	return &providers.NameEnquiryResponse{
		AccountName: "Mock External Account",
		Currency:    money.USD,
	}, nil
}

func (m *mockCurrencyCloudProvider) GetExchangeRate(ctx context.Context, req providers.ExchangeRateRequest) (*providers.ExchangeRateResponse, error) {
	return &providers.ExchangeRateResponse{Rate: 1.2}, nil
}

func (m *mockCurrencyCloudProvider) SendPayout(ctx context.Context, req providers.PayoutRequest) (*providers.PayoutResponse, error) {
	return &providers.PayoutResponse{
		ProviderRef:  "ref_123",
		ProviderName: "currencycloud",
	}, nil
}

func (m *mockCurrencyCloudProvider) SupportsCurrency(currency money.Currency) bool {
	return true
}

func (m *mockCurrencyCloudProvider) Name() string {
	return "currencycloud"
}
