package utils

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrNotAuthorized  = errors.New("not authorized")
	ErrForbidden      = errors.New("forbidden")
	ErrDuplicatedKey  = errors.New("duplicate entity")
	ErrBadRequest     = errors.New("bad request")
	ErrNotImplemented = errors.New("not implemented")
	ErrCacheMiss      = errors.New("cache miss")
	ErrInternal       = errors.New("server error")
)

type WrappedError interface {
	error
	Unwrap() error
	GetMessage() string
}

type wrappedError struct {
	Message string
	Err     error
}

func (e *wrappedError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *wrappedError) Unwrap() error {
	return e.Err
}

func (e *wrappedError) GetMessage() string {
	return e.Message
}

// wrapErrorMessage creates a new wrapped error
func wrapErrorMessage(err error, msg string) WrappedError {
	return &wrappedError{
		Message: msg,
		Err:     err,
	}
}

// IsWrappedError checks if the error implements WrappedError
func IsWrappedError(err error) (WrappedError, bool) {
	var w WrappedError
	ok := errors.As(err, &w)
	return w, ok
}

func NotFoundErr(message string) error {
	return wrapErrorMessage(ErrNotFound, message)
}

func NotAuthorizedErr(message string) error {
	return wrapErrorMessage(ErrNotAuthorized, message)
}

func ForbiddenErr(message string) error {
	return wrapErrorMessage(ErrForbidden, message)
}

func DuplicateKeyErr(message string) error {
	return wrapErrorMessage(ErrDuplicatedKey, message)
}

func BadRequestErr(message string) error {
	return wrapErrorMessage(ErrBadRequest, message)
}

func NotImplementedErr(message string) error {
	return wrapErrorMessage(ErrNotImplemented, message)
}

func ServerErr(err error) error {
	return wrapErrorMessage(ErrInternal, err.Error())
}
