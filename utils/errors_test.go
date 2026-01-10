package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrappedError(t *testing.T) {
	err := NotFoundErr("user not found")

	wrapped, ok := IsWrappedError(err)
	assert.True(t, ok)
	assert.Equal(t, "user not found", wrapped.GetMessage())
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		baseErr error
		message string
	}{
		{
			name:    "NotFoundErr",
			err:     NotFoundErr("resource not found"),
			baseErr: ErrNotFound,
			message: "resource not found",
		},
		{
			name:    "BadRequestErr",
			err:     BadRequestErr("invalid input"),
			baseErr: ErrBadRequest,
			message: "invalid input",
		},
		{
			name:    "NotAuthorizedErr",
			err:     NotAuthorizedErr("unauthorized"),
			baseErr: ErrNotAuthorized,
			message: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, errors.Is(tt.err, tt.baseErr))
			wrapped, ok := IsWrappedError(tt.err)
			assert.True(t, ok)
			assert.Equal(t, tt.message, wrapped.GetMessage())
		})
	}
}
