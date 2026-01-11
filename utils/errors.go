package utils

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrDuplicatedKey = errors.New("duplicate entity")
	ErrBadRequest    = errors.New("bad request")
	ErrInternal      = errors.New("server error")
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

func wrapErrorMessage(err error, msg string) WrappedError {
	return &wrappedError{
		Message: msg,
		Err:     err,
	}
}

func IsWrappedError(err error) (WrappedError, bool) {
	var w WrappedError
	ok := errors.As(err, &w)
	return w, ok
}

func NotFoundErr(message string) error {
	return wrapErrorMessage(ErrNotFound, message)
}

func DuplicateKeyErr(message string) error {
	return wrapErrorMessage(ErrDuplicatedKey, message)
}

func BadRequestErr(message string) error {
	return wrapErrorMessage(ErrBadRequest, message)
}

func ServerErr(err error) error {
	return wrapErrorMessage(ErrInternal, err.Error())
}
