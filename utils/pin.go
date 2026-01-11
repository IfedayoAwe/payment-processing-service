package utils

import (
	"errors"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

const (
	PinMinLength = 4
	PinMaxLength = 4
)

var pinRegex = regexp.MustCompile(`^\d{4}$`)

func HashPIN(pin string) (string, error) {
	if !IsValidPIN(pin) {
		return "", errors.New("invalid PIN format")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash PIN: %w", err)
	}

	return string(hash), nil
}

func VerifyPIN(hashedPIN, pin string) error {
	if hashedPIN == "" {
		return errors.New("PIN not set")
	}

	if !IsValidPIN(pin) {
		return errors.New("invalid PIN format")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPIN), []byte(pin)); err != nil {
		return errors.New("invalid PIN")
	}

	return nil
}

func IsValidPIN(pin string) bool {
	return len(pin) == PinMinLength && pinRegex.MatchString(pin)
}
