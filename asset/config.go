package asset

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
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

type RecordFilter map[string][]string
type ConfigType []Type

type Config struct {
	Text          string
	Types         ConfigType
	Services      []string
	Size          int
	Offset        int
	SortBy        string `validate:"omitempty,oneof=name types services created_at updated_at"`
	SortDirection string `validate:"omitempty,oneof=asc desc"`
	Filters       RecordFilter
	URN           string
	Name          string
}

// Validate will check whether fields in the filter fulfills the constraint
func (cfg *Config) Validate() error {

	err := validate.Struct(cfg)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		var errStrs []string
		for _, e := range errs {
			if e.Tag() == "oneof" {
				errStr := fmt.Sprintf("error filter \"%s\" for key \"%s\" not recognized, only support \"%s\"", e.Value(), e.Field(), e.Param())
				errStrs = append(errStrs, errStr)
				continue
			}
			errStrs = append(errStrs, e.Error())
		}
		return errors.New(strings.Join(errStrs, " and "))
	}
	return err
}

//// AssignDefault will populate default value to filter
//func (f *Filter) AssignDefault() {
//	if len(strings.TrimSpace(f.Type)) == 0 {
//		f.Type = "all"
//	}
//}
