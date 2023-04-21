package utils

import (
	"errors"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
)

var (
	validate *validator.Validate
)

func init() {
	validate = validator.New()
}
func Validate(data interface{}) error {
	errs := validate.Struct(data)
	if errs == nil {
		return nil
	}
	return errors.New(errorData(errs.(validator.ValidationErrors)))
}

func errorData(errs []validator.FieldError) string {
	for _, err := range errs {
		return fmt.Sprintf("%s is %s %s", err.Field(), err.Tag(), err.Param())
	}
	return "unknown error"
}

func RegisterValidate(tag string, fl validator.Func, callValidationEvenIfNull ...bool) error {
	return validate.RegisterValidation(tag, fl, callValidationEvenIfNull...)
}
