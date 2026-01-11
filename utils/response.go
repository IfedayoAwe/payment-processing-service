package utils

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SuccessResponse struct {
	Data    any    `json:"data"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

type InternalErrorResponse struct {
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

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
	case errors.Is(baseErr, ErrDuplicatedKey):
		return Conflict(c, message)
	case errors.Is(baseErr, ErrBadRequest):
		return BadRequest(c, message)
	case errors.Is(baseErr, ErrInternal):
		fallthrough
	default:
		return InternalError(c, message)
	}
}

func Success(c echo.Context, data any, message string) error {
	return c.JSON(http.StatusOK, SuccessResponse{
		Data:    data,
		Message: message,
	})
}

func Created(c echo.Context, data any, message string) error {
	return c.JSON(http.StatusCreated, SuccessResponse{
		Data:    data,
		Message: message,
	})
}

func BadRequest(c echo.Context, message string) error {
	return errorResponse(c, http.StatusBadRequest, message)
}

func NotFound(c echo.Context, message string) error {
	return errorResponse(c, http.StatusNotFound, message)
}

func Conflict(c echo.Context, message string) error {
	return errorResponse(c, http.StatusConflict, message)
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

func ValidationError(c echo.Context, errors map[string]string) error {
	return c.JSON(http.StatusBadRequest, ValidationErrorResponse{
		Message: "Validation failed",
		Errors:  errors,
	})
}
