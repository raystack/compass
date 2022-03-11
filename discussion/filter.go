package discussion

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

type Filter struct {
	Type          string `json:"type" validate:"omitempty,oneof=openended issues qanda all"`
	State         string `json:"state" validate:"omitempty,oneof=open closed"`
	Assignees     []string
	Assets        []string
	Owner         string
	Labels        []string
	SortBy        string `json:"sort" validate:"omitempty,oneof=created_at updated_at"`
	SortDirection string `json:"direction" validate:"omitempty,oneof=asc desc"`
	Size          int    `json:"size" validate:"omitempty,gte=0"`
	Offset        int    `json:"offset" validate:"omitempty,gte=0"`
}

// Validate will check whether fields in the filter fulfills the constraint
func (f *Filter) Validate() error {

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
