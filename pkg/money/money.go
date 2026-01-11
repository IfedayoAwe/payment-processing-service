package money

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
)

func (c Currency) String() string {
	return string(c)
}

func (c Currency) IsValid() bool {
	return c == USD || c == EUR || c == GBP
}

func ParseCurrency(s string) (Currency, error) {
	c := Currency(s)
	if !c.IsValid() {
		return "", fmt.Errorf("invalid currency: %s", s)
	}
	return c, nil
}

type Money struct {
	Amount   int64    `json:"amount"`
	Currency Currency `json:"currency"`
}

func NewMoney(amount int64, currency Currency) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.Currency, float64(m.Amount)/100.0)
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s + %s", m.Currency, other.Currency)
	}
	return NewMoney(m.Amount+other.Amount, m.Currency), nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s - %s", m.Currency, other.Currency)
	}
	if m.Amount < other.Amount {
		return Money{}, fmt.Errorf("insufficient funds: %d < %d", m.Amount, other.Amount)
	}
	return NewMoney(m.Amount-other.Amount, m.Currency), nil
}

func (m Money) IsPositive() bool {
	return m.Amount > 0
}

func (m Money) IsZero() bool {
	return m.Amount == 0
}

func (m Money) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Money) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

func FromMajorUnits(amount float64, currency Currency) Money {
	return NewMoney(int64(amount*100), currency)
}

func ToMajorUnits(amount int64) float64 {
	return float64(amount) / 100.0
}
