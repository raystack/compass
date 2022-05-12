package tag

import (
	"fmt"
)

type NotFoundError struct {
	AssetID  string
	Template string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf(
		"could not find tag with asset id: \"%s\", template: \"%s\"",
		e.AssetID,
		e.Template,
	)
}

type TemplateNotFoundError struct {
	URN string
}

func (e TemplateNotFoundError) Error() string {
	return fmt.Sprintf("could not find template \"%s\"", e.URN)
}

type DuplicateError struct {
	AssetID     string
	TemplateURN string
}

func (e DuplicateError) Error() string {
	return fmt.Sprintf("tag of asset ID \"%s\" and template URN \"%s\" already exists", e.AssetID, e.TemplateURN)
}

type DuplicateTemplateError struct {
	URN string
}

func (e DuplicateTemplateError) Error() string {
	return fmt.Sprintf("template \"%s\" already exists", e.URN)
}

type ValidationError struct {
	Err error
}

func (e ValidationError) Error() string {
	return e.Err.Error()
}
