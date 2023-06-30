package tag

import (
	"fmt"

	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"
	"github.com/raystack/compass/core/tag/validator"
)

// newValidator initializes validator for tag
func newValidator() validator.Validator {
	v, err := validator.NewBuilder().
		WithTranslations([]validator.Translation{
			{
				Tag:      "required",
				Message:  "cannot be empty",
				Override: true,
			},
			{
				Tag:      "min",
				Message:  "must be at least {0}",
				Override: true,
				TranslationFunc: func(t ut.Translator, fe v.FieldError) string {
					output, _ := t.T(fe.Tag(), fe.Param())
					return output
				},
			},
		}).
		Build()
	if err != nil {
		panic(err)
	}

	return v
}

// newTemplateValidator initializes validator for tag template
func newTemplateValidator() validator.Validator {
	v, err := validator.NewBuilder().
		WithStructValidations([]validator.StructValidation{
			{
				Type: Template{},
				Func: func(sl v.StructLevel) {
					template, ok := sl.Current().Interface().(Template)
					if !ok {
						sl.ReportError(nil, "struct", "", "is_not_a_template", "")
					}

					for i, field := range template.Fields {
						if field.DataType == "enumerated" {
							if len(field.Options) == 0 {
								sl.ReportError(
									nil, fmt.Sprintf("fields[%d].options", i), "", "enumerated_restricted", "",
								)
							}
							for _, opt := range field.Options {
								if opt == "" {
									sl.ReportError(
										nil, fmt.Sprintf("fields[%d].options", i), "", "element_not_empty", "",
									)
								}
							}
						}
					}
				},
			},
		}).
		WithTranslations([]validator.Translation{
			{
				Tag:      "required",
				Message:  "cannot be empty",
				Override: true,
			},
			{
				Tag:      "min",
				Message:  "must be at least {0}",
				Override: true,
				TranslationFunc: func(t ut.Translator, fe v.FieldError) string {
					output, _ := t.T(fe.Tag(), fe.Param())
					return output
				},
			},
			{
				Tag:     "enumerated_restricted",
				Message: "cannot be empty with data_type [enumerated]",
				TranslationFunc: func(t ut.Translator, fe v.FieldError) string {
					output, _ := t.T(fe.Tag())
					return output
				},
			},
			{
				Tag:     "element_not_empty",
				Message: "cannot contain empty element",
				TranslationFunc: func(t ut.Translator, fe v.FieldError) string {
					output, _ := t.T(fe.Tag())
					return output
				},
			},
		}).
		Build()
	if err != nil {
		panic(err)
	}

	return v
}
