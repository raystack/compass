package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// singleton
var validate *validator.Validate

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}

func ValidateStruct(f interface{}) error {
	err := getValidator().Struct(f)
	return checkError(err)
}

func ValidateOneOf(value string, enums ...string) error {
	tags := "omitempty,oneof=" + strings.Join(enums, " ")
	err := getValidator().Var(value, tags)
	return checkError(err)
}

func getValidator() *validator.Validate {
	if validate == nil {
		validate = newValidator()
	}
	return validate
}

func checkError(err error) error {
	if err == nil {
		return nil
	}

	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		return err
	}

	var errStrs []string
	for _, e := range errs {
		switch e.Tag() {
		case "oneof":
			errStr := fmt.Sprintf("error value %q", e.Value())
			if e.Field() != "" {
				errStr += fmt.Sprintf(" for key %q", e.Field())
			}
			errStr += fmt.Sprintf(" not recognized, only support %q", e.Param())
			errStrs = append(errStrs, errStr)

		case "gte":
			errStrs = append(errStrs, fmt.Sprintf("%s cannot be less than %s", e.Field(), e.Param()))

		default:
			errStrs = append(errStrs, e.Error())
		}
	}
	return errors.New(strings.Join(errStrs, " and "))
}
