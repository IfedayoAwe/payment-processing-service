package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeCursor(t *testing.T) {
	t.Run("valid cursor", func(t *testing.T) {
		now := time.Now()
		id := "tx_123"

		encoded := EncodeCursor(now, id)
		assert.NotEmpty(t, encoded)
	})

	t.Run("empty id", func(t *testing.T) {
		now := time.Now()
		encoded := EncodeCursor(now, "")
		assert.NotEmpty(t, encoded)
	})
}

func TestDecodeCursor(t *testing.T) {
	t.Run("valid cursor round trip", func(t *testing.T) {
		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		id := "tx_12345"

		encoded := EncodeCursor(now, id)
		decoded, err := DecodeCursor(encoded)

		require.NoError(t, err)
		require.NotNil(t, decoded)
		assert.Equal(t, now.Unix(), decoded.CreatedAt.Unix())
		assert.Equal(t, id, decoded.ID)
	})

	t.Run("empty cursor string", func(t *testing.T) {
		decoded, err := DecodeCursor("")
		assert.NoError(t, err)
		assert.Nil(t, decoded)
	})

	t.Run("invalid base64", func(t *testing.T) {
		decoded, err := DecodeCursor("not-valid-base64!!!")
		assert.Error(t, err)
		assert.Nil(t, decoded)
	})

	t.Run("invalid json", func(t *testing.T) {
		invalidBase64 := "dGhpcyBpcyBub3QgdmFsaWQganNvbg=="
		decoded, err := DecodeCursor(invalidBase64)
		assert.Error(t, err)
		assert.Nil(t, decoded)
	})
}

func TestCursorRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		id   string
	}{
		{"simple id", "tx_1"},
		{"uuid-like id", "550e8400-e29b-41d4-a716-446655440000"},
		{"numeric id", "123456"},
		{"empty id", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now().Truncate(time.Second)
			encoded := EncodeCursor(now, tc.id)
			decoded, err := DecodeCursor(encoded)

			require.NoError(t, err)
			require.NotNil(t, decoded)
			assert.WithinDuration(t, now, decoded.CreatedAt, time.Second)
			assert.Equal(t, tc.id, decoded.ID)
		})
	}
}
