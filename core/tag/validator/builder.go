package validator

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translation "github.com/go-playground/validator/v10/translations/en"
)

const defaultLocale = "en"

// Builder is type to build validator
type Builder struct {
	fieldValidations  []FieldValidation
	structValidations []StructValidation
	translations      []Translation

	validate   *validator.Validate
	translator ut.Translator
}

// WithTranslations tells builder to include custom translation
func (b *Builder) WithTranslations(translations []Translation) *Builder {
	output := *b
	output.translations = translations
	return &output
}

// WithStructValidations tells builder to include custom struct validation
func (b *Builder) WithStructValidations(structValidations []StructValidation) *Builder {
	output := *b
	output.structValidations = structValidations
	return &output
}

// WithFieldValidations tells builder to include custom field validation
func (b *Builder) WithFieldValidations(fieldValidations []FieldValidation) *Builder {
	output := *b
	output.fieldValidations = fieldValidations
	return &output
}

// Build builds the validator
func (b *Builder) Build() (Validator, error) {
	b.translator = b.initializeTranslator()
	validate, err := b.initializeValidate(b.translator)
	if err != nil {
		return nil, err
	}
	b.validate = validate

	if b.fieldValidations != nil {
		err = b.registerFieldValidations(b.validate, b.fieldValidations)
		if err != nil {
			return nil, err
		}
	}
	if b.structValidations != nil {
		b.registerStructValidations(b.validate, b.structValidations)
	}
	if b.translations != nil {
		err = b.registerTranslations(b.validate, b.translator, b.translations)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

// Validate validates the type for any violation
func (b *Builder) Validate(s interface{}) error {
	err := b.validate.Struct(s)
	if err == nil {
		return nil
	}
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return err
	}
	fieldErrors := make(FieldError)
	for _, f := range validationErrs {
		var field string
		splitNamespace := strings.Split(f.Namespace(), ".")
		if len(splitNamespace) > 0 {
			if len(splitNamespace) > 1 {
				field = strings.Join(splitNamespace[1:], ".")
			} else {
				field = splitNamespace[0]
			}
		}

		errMsg := f.Translate(b.translator)
		fieldErrors[field] = errMsg
	}
	return fieldErrors
}

func (b *Builder) initializeTranslator() ut.Translator {
	universalTranslator := ut.New(en.New(), en.New())
	translator, _ := universalTranslator.GetTranslator(defaultLocale)
	return translator
}

func (b *Builder) initializeValidate(translator ut.Translator) (*validator.Validate, error) {
	validate := validator.New()
	validate.RegisterTagNameFunc(b.tagNameFunc)
	if err := en_translation.RegisterDefaultTranslations(validate, translator); err != nil {
		return nil, err
	}
	return validate, nil
}

func (b *Builder) registerFieldValidations(validate *validator.Validate, fieldValidations []FieldValidation) error {
	for _, f := range fieldValidations {
		err := validate.RegisterValidation(f.Tag, f.Func)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) registerStructValidations(validate *validator.Validate, structValidations []StructValidation) {
	for _, s := range structValidations {
		validate.RegisterStructValidation(s.Func, s.Type)
	}
}

func (b *Builder) registerTranslations(validate *validator.Validate, translator ut.Translator, translations []Translation) error {
	for _, t := range translations {
		registerFn := getRegisterFn(t.Tag, t.Message, t.Override)
		transFunc := t.TranslationFunc
		if transFunc == nil {
			transFunc = b.transFunc
		}
		err := validate.RegisterTranslation(t.Tag, translator, registerFn, transFunc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) tagNameFunc(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func (b *Builder) transFunc(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T(fe.Tag())
	return t
}

func getRegisterFn(tag, translation string, override bool) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) error {
		return ut.Add(tag, translation, override)
	}
}

// NewBuilder initializes builder
func NewBuilder() *Builder {
	return &Builder{}
}
