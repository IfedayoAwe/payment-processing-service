package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCurrency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Currency
		wantErr  bool
	}{
		{"valid USD", "USD", USD, false},
		{"valid EUR", "EUR", EUR, false},
		{"valid GBP", "GBP", GBP, false},
		{"invalid currency", "INVALID", "", true},
		{"empty string", "", "", true},
		{"lowercase", "usd", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCurrency(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCurrency_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		expected bool
	}{
		{"USD is valid", USD, true},
		{"EUR is valid", EUR, true},
		{"GBP is valid", GBP, true},
		{"invalid currency", Currency("INVALID"), false},
		{"empty currency", Currency(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.currency.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_IsPositive(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected bool
	}{
		{"positive amount", NewMoney(100, USD), true},
		{"zero amount", NewMoney(0, USD), false},
		{"negative amount", NewMoney(-100, USD), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.IsPositive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected bool
	}{
		{"zero amount", NewMoney(0, USD), true},
		{"positive amount", NewMoney(100, USD), false},
		{"negative amount", NewMoney(-100, USD), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.IsZero()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_Add(t *testing.T) {
	t.Run("same currency", func(t *testing.T) {
		m1 := NewMoney(100, USD)
		m2 := NewMoney(50, USD)

		result, err := m1.Add(m2)
		require.NoError(t, err)
		assert.Equal(t, int64(150), result.Amount)
		assert.Equal(t, USD, result.Currency)
	})

	t.Run("different currencies", func(t *testing.T) {
		m1 := NewMoney(100, USD)
		m2 := NewMoney(50, EUR)

		result, err := m1.Add(m2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot add different currencies")
		assert.Zero(t, result.Amount)
	})
}

func TestMoney_Subtract(t *testing.T) {
	t.Run("same currency sufficient funds", func(t *testing.T) {
		m1 := NewMoney(100, USD)
		m2 := NewMoney(30, USD)

		result, err := m1.Subtract(m2)
		require.NoError(t, err)
		assert.Equal(t, int64(70), result.Amount)
		assert.Equal(t, USD, result.Currency)
	})

	t.Run("different currencies", func(t *testing.T) {
		m1 := NewMoney(100, USD)
		m2 := NewMoney(50, EUR)

		result, err := m1.Subtract(m2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot subtract different currencies")
		assert.Zero(t, result.Amount)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		m1 := NewMoney(50, USD)
		m2 := NewMoney(100, USD)

		result, err := m1.Subtract(m2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
		assert.Zero(t, result.Amount)
	})
}

func TestFromMajorUnits(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency Currency
		expected int64
	}{
		{"dollars to cents", 10.50, USD, 1050},
		{"whole dollars", 100.0, USD, 10000},
		{"small amount", 0.01, USD, 1},
		{"euros", 25.75, EUR, 2575},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromMajorUnits(tt.amount, tt.currency)
			assert.Equal(t, tt.expected, result.Amount)
			assert.Equal(t, tt.currency, result.Currency)
		})
	}
}

func TestToMajorUnits(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		expected float64
	}{
		{"cents to dollars", 1050, 10.50},
		{"whole dollars", 10000, 100.0},
		{"single cent", 1, 0.01},
		{"zero", 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToMajorUnits(tt.amount)
			assert.Equal(t, tt.expected, result)
		})
	}
}
