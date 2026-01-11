package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
)

type CurrencyCloudProvider struct {
	name string
}

func NewCurrencyCloudProvider() *CurrencyCloudProvider {
	return &CurrencyCloudProvider{name: "CurrencyCloud"}
}

func (p *CurrencyCloudProvider) Name() string {
	return p.name
}

func (p *CurrencyCloudProvider) SupportsCurrency(currency money.Currency) bool {
	return currency == money.USD || currency == money.EUR
}

func (p *CurrencyCloudProvider) SendPayout(ctx context.Context, req PayoutRequest) (*PayoutResponse, error) {
	time.Sleep(30 * time.Millisecond)

	var ref ProviderReference
	if req.ProviderRef != nil {
		ref = *req.ProviderRef
	} else {
		ref = ProviderReference(fmt.Sprintf("CC-%d", time.Now().UnixNano()))
	}

	return &PayoutResponse{
		ProviderRef: ref,
		Status:      "pending",
	}, nil
}

func (p *CurrencyCloudProvider) NameEnquiry(ctx context.Context, req NameEnquiryRequest) (*NameEnquiryResponse, error) {
	time.Sleep(50 * time.Millisecond)

	if len(req.AccountNumber) < 4 {
		return &NameEnquiryResponse{
			AccountName: "Mock Account Holder",
			Currency:    money.USD,
		}, nil
	}

	lastFour := req.AccountNumber[len(req.AccountNumber)-4:]
	return &NameEnquiryResponse{
		AccountName: fmt.Sprintf("Mock Account Holder %s", lastFour),
		Currency:    money.USD,
	}, nil
}

func (p *CurrencyCloudProvider) GetExchangeRate(ctx context.Context, req ExchangeRateRequest) (*ExchangeRateResponse, error) {
	time.Sleep(30 * time.Millisecond)

	if req.FromCurrency == req.ToCurrency {
		return &ExchangeRateResponse{Rate: 1.0}, nil
	}

	rates := map[string]map[string]float64{
		"USD": {"EUR": 0.85, "GBP": 0.75},
		"EUR": {"USD": 1.18, "GBP": 0.88},
		"GBP": {"USD": 1.33, "EUR": 1.14},
	}

	fromRates, ok := rates[req.FromCurrency.String()]
	if !ok {
		return nil, fmt.Errorf("unsupported from currency: %s", req.FromCurrency)
	}

	rate, ok := fromRates[req.ToCurrency.String()]
	if !ok {
		return nil, fmt.Errorf("unsupported to currency: %s", req.ToCurrency)
	}

	return &ExchangeRateResponse{Rate: rate}, nil
}
