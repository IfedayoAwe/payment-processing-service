package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidPIN(t *testing.T) {
	tests := []struct {
		name     string
		pin      string
		expected bool
	}{
		{"valid 5 digit pin", "12345", true},
		{"valid pin with zeros", "00000", true},
		{"too short", "1234", false},
		{"too long", "123456", false},
		{"contains letters", "12ab5", false},
		{"contains special chars", "12@45", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPIN(tt.pin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashPIN(t *testing.T) {
	t.Run("valid pin", func(t *testing.T) {
		hash, err := HashPIN("12345")
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, "12345", hash)
	})

	t.Run("invalid pin format", func(t *testing.T) {
		hash, err := HashPIN("1234")
		assert.Error(t, err)
		assert.Empty(t, hash)
		assert.Contains(t, err.Error(), "invalid PIN format")
	})

	t.Run("same pin produces different hashes", func(t *testing.T) {
		hash1, err1 := HashPIN("12345")
		require.NoError(t, err1)

		hash2, err2 := HashPIN("12345")
		require.NoError(t, err2)

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestVerifyPIN(t *testing.T) {
	t.Run("correct pin", func(t *testing.T) {
		hash, err := HashPIN("12345")
		require.NoError(t, err)

		err = VerifyPIN(hash, "12345")
		assert.NoError(t, err)
	})

	t.Run("incorrect pin", func(t *testing.T) {
		hash, err := HashPIN("12345")
		require.NoError(t, err)

		err = VerifyPIN(hash, "56789")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid PIN")
	})

	t.Run("empty hash", func(t *testing.T) {
		err := VerifyPIN("", "12345")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PIN not set")
	})

	t.Run("invalid pin format", func(t *testing.T) {
		hash, err := HashPIN("12345")
		require.NoError(t, err)

		err = VerifyPIN(hash, "1234")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid PIN format")
	})
}
