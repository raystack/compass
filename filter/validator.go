package filter

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

func Validate(f interface{}) error {
	if validate == nil {
		validate = newValidator()
	}
	err := validate.Struct(f)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		errStrs := []string{}
		for _, e := range errs {
			if e.Tag() == "oneof" {
				errStr := fmt.Sprintf("error filter \"%s\" for key \"%s\" not recognized, only support \"%s\"", e.Value(), e.Field(), e.Param())
				errStrs = append(errStrs, errStr)
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
	return err
}
