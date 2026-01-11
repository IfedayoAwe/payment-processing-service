package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Email  string `validate:"required,email"`
	Amount int    `validate:"required,min=1"`
}

func TestFormatValidationErrors(t *testing.T) {
	validate := InitValidator()
	testStruct := TestStruct{
		Email:  "invalid-email",
		Amount: 0,
	}

	err := validate.Struct(testStruct)
	assert.Error(t, err)

	errors := FormatValidationErrors(err)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "Email")
	assert.Contains(t, errors, "Amount")
}

func TestInitValidator(t *testing.T) {
	v1 := InitValidator()
	v2 := InitValidator()

	// Should return the same instance
	assert.Equal(t, v1, v2)
}

func TestGetTranslator(t *testing.T) {
	t1 := GetTranslator()
	t2 := GetTranslator()

	// Should return the same instance
	assert.Equal(t, t1, t2)
	assert.NotNil(t, t1)
}
