package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/IfedayoAwe/payment-processing-service/utils"
)

type NameEnquiryService interface {
	EnquireAccountName(ctx context.Context, accountNumber, bankCode string) (*NameEnquiryResult, error)
}

type NameEnquiryResult struct {
	AccountName string
	IsInternal  bool
	Currency    money.Currency
}

type nameEnquiryService struct {
	queries  *gen.Queries
	provider *providers.Processor
}

func newNameEnquiryService(queries *gen.Queries, provider *providers.Processor) NameEnquiryService {
	return &nameEnquiryService{
		queries:  queries,
		provider: provider,
	}
}

func (nes *nameEnquiryService) EnquireAccountName(ctx context.Context, accountNumber, bankCode string) (*NameEnquiryResult, error) {
	if accountNumber == "" {
		return nil, utils.BadRequestErr("account_number is required")
	}
	if bankCode == "" {
		return nil, utils.BadRequestErr("bank_code is required")
	}

	bankAccount, err := nes.queries.GetBankAccountByAccountAndBankCode(ctx, gen.GetBankAccountByAccountAndBankCodeParams{
		AccountNumber: accountNumber,
		BankCode:      bankCode,
	})
	if err == nil {
		currency, parseErr := money.ParseCurrency(bankAccount.Currency)
		if parseErr != nil {
			return nil, utils.ServerErr(fmt.Errorf("invalid currency in bank account: %w", parseErr))
		}

		accountName := ""
		if bankAccount.AccountName.Valid {
			accountName = bankAccount.AccountName.String
		}

		return &NameEnquiryResult{
			AccountName: accountName,
			IsInternal:  true,
			Currency:    currency,
		}, nil
	}

	if err != sql.ErrNoRows {
		return nil, utils.ServerErr(fmt.Errorf("check internal account: %w", err))
	}

	resp, err := nes.provider.NameEnquiry(ctx, providers.NameEnquiryRequest{
		AccountNumber: accountNumber,
		BankCode:      bankCode,
	})
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("external name enquiry failed: %w", err))
	}

	return &NameEnquiryResult{
		AccountName: resp.AccountName,
		IsInternal:  false,
		Currency:    resp.Currency,
	}, nil
}
