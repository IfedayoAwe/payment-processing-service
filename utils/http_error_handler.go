package utils

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func HTTPErrorHandler(err error, c echo.Context) {
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		errors := FormatValidationErrors(validationErr)
		_ = ValidationError(c, errors)
		return
	}

	he, ok := err.(*echo.HTTPError)
	if ok {
		if innerErr, ok := he.Internal.(validator.ValidationErrors); ok {
			errors := FormatValidationErrors(innerErr)
			_ = ValidationError(c, errors)
			return
		}

		if msgErr, ok := he.Message.(error); ok {
			if validationErr, ok := msgErr.(validator.ValidationErrors); ok {
				errors := FormatValidationErrors(validationErr)
				_ = ValidationError(c, errors)
				return
			}
		}

		message := http.StatusText(he.Code)
		if msg, ok := he.Message.(string); ok {
			message = msg
		}
		_ = c.JSON(he.Code, ErrorResponse{Message: message})
		return
	}

	_ = HandleError(c, err)
}
