package utils

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// HTTPErrorHandler handles validation errors and formats them properly
func HTTPErrorHandler(err error, c echo.Context) {
	// Check if it's a validation error directly
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		errors := FormatValidationErrors(validationErr)
		_ = ValidationError(c, errors)
		return
	}

	// Check if it's an Echo HTTP error (validation errors are often wrapped)
	he, ok := err.(*echo.HTTPError)
	if ok {
		// Check if the inner error is a validation error
		if innerErr, ok := he.Internal.(validator.ValidationErrors); ok {
			errors := FormatValidationErrors(innerErr)
			_ = ValidationError(c, errors)
			return
		}

		// Check if the message itself is a validation error
		if msgErr, ok := he.Message.(error); ok {
			if validationErr, ok := msgErr.(validator.ValidationErrors); ok {
				errors := FormatValidationErrors(validationErr)
				_ = ValidationError(c, errors)
				return
			}
		}

		// Regular HTTP error
		message := http.StatusText(he.Code)
		if msg, ok := he.Message.(string); ok {
			message = msg
		}
		_ = c.JSON(he.Code, ErrorResponse{Message: message})
		return
	}

	// Default error handling
	_ = HandleError(c, err)
}
