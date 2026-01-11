package providers

import (
	"context"
	"fmt"

	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
)

type Processor struct {
	payoutProviders       []PayoutProvider
	nameEnquiryProviders  []NameEnquiryProvider
	exchangeRateProviders []ExchangeRateProvider
}

func NewProcessor() *Processor {
	return &Processor{
		payoutProviders:       []PayoutProvider{},
		nameEnquiryProviders:  []NameEnquiryProvider{},
		exchangeRateProviders: []ExchangeRateProvider{},
	}
}

func (p *Processor) RegisterPayoutProvider(provider PayoutProvider) {
	p.payoutProviders = append(p.payoutProviders, provider)
}

func (p *Processor) RegisterNameEnquiryProvider(provider NameEnquiryProvider) {
	p.nameEnquiryProviders = append(p.nameEnquiryProviders, provider)
}

func (p *Processor) RegisterExchangeRateProvider(provider ExchangeRateProvider) {
	p.exchangeRateProviders = append(p.exchangeRateProviders, provider)
}

func (p *Processor) SelectProvider(currency money.Currency) (PayoutProvider, error) {
	for _, provider := range p.payoutProviders {
		if provider.SupportsCurrency(currency) {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("no provider available for currency: %s", currency)
}

func (p *Processor) SendPayout(ctx context.Context, req PayoutRequest) (*PayoutResponse, error) {
	provider, err := p.SelectProvider(req.Amount.Currency)
	if err != nil {
		return nil, err
	}

	resp, err := provider.SendPayout(ctx, req)
	if err != nil {
		return nil, err
	}

	resp.ProviderName = provider.Name()
	return resp, nil
}

func (p *Processor) NameEnquiry(ctx context.Context, req NameEnquiryRequest) (*NameEnquiryResponse, error) {
	if len(p.nameEnquiryProviders) == 0 {
		return nil, fmt.Errorf("no name enquiry providers available")
	}

	return p.nameEnquiryProviders[0].NameEnquiry(ctx, req)
}

func (p *Processor) GetExchangeRate(ctx context.Context, req ExchangeRateRequest) (*ExchangeRateResponse, error) {
	if len(p.exchangeRateProviders) == 0 {
		return nil, fmt.Errorf("no exchange rate providers available")
	}

	return p.exchangeRateProviders[0].GetExchangeRate(ctx, req)
}
