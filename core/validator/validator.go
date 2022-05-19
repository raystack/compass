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
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return validate
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
	if err != nil {
		errs := err.(validator.ValidationErrors)
		errStrs := []string{}
		for _, e := range errs {
			if e.Tag() == "oneof" {
				errStrValue := fmt.Sprintf("error value \"%s\"", e.Value())
				if e.Field() != "" {
					errStrValue = errStrValue + fmt.Sprintf(" for key \"%s\"", e.Field())
				}
				errStrValue = errStrValue + fmt.Sprintf(" not recognized, only support \"%s\"", e.Param())
				errStrs = append(errStrs, errStrValue)
				continue
			}

			if e.Tag() == "gte" {
				errStrs = append(errStrs, fmt.Sprintf("%s cannot be less than %s", e.Field(), e.Param()))
				continue
			}

			errStrs = append(errStrs, e.Error())
		}
		return errors.New(strings.Join(errStrs, " and "))
	}
	return nil
}
