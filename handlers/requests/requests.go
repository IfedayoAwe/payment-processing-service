package requests

import (
	"fmt"

	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/utils"
)

type AmountRequest struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Currency string  `json:"currency" validate:"required,oneof=USD EUR GBP"`
}

func (a *AmountRequest) ToMoney() (money.Money, error) {
	currency, err := money.ParseCurrency(a.Currency)
	if err != nil {
		return money.Money{}, fmt.Errorf("invalid currency: %w", err)
	}

	return money.FromMajorUnits(a.Amount, currency), nil
}

type CreateInternalTransferRequest struct {
	FromCurrency    string        `json:"from_currency" validate:"required,oneof=USD EUR GBP"`
	ToAccountNumber string        `json:"to_account_number" validate:"required"`
	ToBankCode      string        `json:"to_bank_code" validate:"required"`
	Amount          AmountRequest `json:"amount" validate:"required"`
}

type CreateExternalTransferRequest struct {
	BankAccountID string        `json:"bank_account_id" validate:"required"`
	FromCurrency  string        `json:"from_currency" validate:"required,oneof=USD EUR GBP"`
	Amount        AmountRequest `json:"amount" validate:"required"`
}

type NameEnquiryRequest struct {
	AccountNumber string `json:"account_number" validate:"required"`
	BankCode      string `json:"bank_code" validate:"required"`
}

type ConfirmTransactionRequest struct {
	PIN string `json:"pin" validate:"required"`
}

func (c *ConfirmTransactionRequest) Validate() error {
	if !utils.IsValidPIN(c.PIN) {
		return fmt.Errorf("PIN must be exactly 4 numeric digits")
	}
	return nil
}
