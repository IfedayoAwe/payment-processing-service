package providers

import (
	"context"
	"fmt"

	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
)

type Processor struct {
	providers []PayoutProvider
}

func NewProcessor() *Processor {
	return &Processor{providers: []PayoutProvider{}}
}

func (p *Processor) RegisterPayoutProvider(provider PayoutProvider) {
	p.providers = append(p.providers, provider)
}

func (p *Processor) SelectProvider(currency money.Currency) (PayoutProvider, error) {
	for _, provider := range p.providers {
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
	if len(p.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	return p.providers[0].NameEnquiry(ctx, req)
}

func (p *Processor) GetExchangeRate(ctx context.Context, req ExchangeRateRequest) (*ExchangeRateResponse, error) {
	if len(p.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	return p.providers[0].GetExchangeRate(ctx, req)
}
