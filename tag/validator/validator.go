package validator

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// Validator is contract to do validation
type Validator interface {
	Validate(interface{}) error
}

// Translation is a type to describe how to translate
type Translation struct {
	Tag             string
	Message         string
	Override        bool
	TranslationFunc func(ut.Translator, validator.FieldError) string
}

// StructValidation is a type to describe how to validate a struct
type StructValidation struct {
	Type interface{}
	Func validator.StructLevelFunc
}

// FieldValidation is a type to describe how to validate a field
type FieldValidation struct {
	Tag  string
	Func validator.Func
}
