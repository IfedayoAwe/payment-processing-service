package utils

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// SuccessResponse defines the standard shape for OK responses.
type SuccessResponse struct {
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// ErrorResponse is the standard shape for all error payloads.
type ErrorResponse struct {
	Message string `json:"message"`
}

// ValidationErrorResponse is the shape for validation error payloads with field-level errors.
type ValidationErrorResponse struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

// InternalErrorResponse is the standard shape for all error payloads.
type InternalErrorResponse struct {
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

// HandleError is the central error handler for all routes.
func HandleError(c echo.Context, err error) error {
	if err == nil {
		return nil
	}

	var (
		baseErr error
		message string
	)

	if wrappedErr, ok := IsWrappedError(err); ok {
		message = wrappedErr.GetMessage()
		baseErr = wrappedErr.Unwrap()
	} else {
		message = err.Error()
		baseErr = err
	}

	switch {
	case errors.Is(baseErr, ErrNotFound):
		return NotFound(c, message)
	case errors.Is(baseErr, ErrNotAuthorized):
		return Unauthorized(c, message)
	case errors.Is(baseErr, ErrForbidden):
		return Forbidden(c, message)
	case errors.Is(baseErr, ErrDuplicatedKey):
		return Conflict(c, message)
	case errors.Is(baseErr, ErrBadRequest):
		return BadRequest(c, message)
	case errors.Is(baseErr, ErrNotImplemented):
		return NotImplemented(c, message)
	case errors.Is(baseErr, ErrInternal):
		fallthrough
	default:
		return InternalError(c, message)
	}
}

// Success Response
func Success(c echo.Context, data any, message string) error {
	return c.JSON(http.StatusOK, SuccessResponse{
		Data:    data,
		Message: message,
	})
}

// Created Response
func Created(c echo.Context, data any, message string) error {
	return c.JSON(http.StatusCreated, SuccessResponse{
		Data:    data,
		Message: message,
	})
}

// ------ Error responses ------

// BadRequest Error responses
func BadRequest(c echo.Context, message string) error {
	return errorResponse(c, http.StatusBadRequest, message)
}

func Unauthorized(c echo.Context, message string) error {
	return errorResponse(c, http.StatusUnauthorized, message)
}

func Forbidden(c echo.Context, message string) error {
	return errorResponse(c, http.StatusForbidden, message)
}

func NotFound(c echo.Context, message string) error {
	return errorResponse(c, http.StatusNotFound, message)
}

func Conflict(c echo.Context, message string) error {
	return errorResponse(c, http.StatusConflict, message)
}

func NotImplemented(c echo.Context, message string) error {
	return errorResponse(c, http.StatusNotImplemented, message)
}

func InternalError(c echo.Context, err string) error {
	return c.JSON(http.StatusInternalServerError, InternalErrorResponse{
		Message: "internal error",
		Detail:  err,
	})
}

func errorResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, ErrorResponse{
		Message: message,
	})
}

// ValidationError returns a validation error response with field-level errors
func ValidationError(c echo.Context, errors map[string]string) error {
	return c.JSON(http.StatusBadRequest, ValidationErrorResponse{
		Message: "Validation failed",
		Errors:  errors,
	})
}
