package providers

import (
	"context"

	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
)

type ProviderReference string

type PayoutRequest struct {
	Amount      money.Money
	Destination BankAccount
	Metadata    map[string]string
	ProviderRef *ProviderReference
}

type BankAccount struct {
	BankName      string
	BankCode      string
	AccountNumber string
	AccountName   string
	Currency      money.Currency
}

type PayoutResponse struct {
	ProviderName string
	ProviderRef  ProviderReference
	Status       string
}

type NameEnquiryRequest struct {
	AccountNumber string
	BankCode      string
}

type NameEnquiryResponse struct {
	AccountName string
	Currency    money.Currency
}

type Provider interface {
	Name() string
}

type ExchangeRateRequest struct {
	FromCurrency money.Currency
	ToCurrency   money.Currency
}

type ExchangeRateResponse struct {
	Rate float64
}

type PayoutProvider interface {
	Provider
	SendPayout(ctx context.Context, req PayoutRequest) (*PayoutResponse, error)
	SupportsCurrency(currency money.Currency) bool
}

type NameEnquiryProvider interface {
	Provider
	NameEnquiry(ctx context.Context, req NameEnquiryRequest) (*NameEnquiryResponse, error)
}

type ExchangeRateProvider interface {
	Provider
	GetExchangeRate(ctx context.Context, req ExchangeRateRequest) (*ExchangeRateResponse, error)
}
