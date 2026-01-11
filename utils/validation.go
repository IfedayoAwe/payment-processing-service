package utils

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	translator ut.Translator
	validate   *validator.Validate
)

func InitValidator() *validator.Validate {
	if validate == nil {
		validate = validator.New()

		english := en.New()
		uni := ut.New(english, english)
		translator, _ = uni.GetTranslator("en")

		_ = enTranslations.RegisterDefaultTranslations(validate, translator)
	}

	return validate
}

func GetTranslator() ut.Translator {
	if translator == nil {
		InitValidator()
	}
	return translator
}

func FormatValidationErrors(err error) map[string]string {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{
			"error": err.Error(),
		}
	}

	trans := GetTranslator()
	errors := make(map[string]string)

	for _, e := range validationErrors {
		errors[e.Field()] = e.Translate(trans)
	}

	return errors
}
