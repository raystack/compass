package tag

import (
	"fmt"
)

type NotFoundError struct {
	URN      string
	Type     string
	Template string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf(
		"could not find tag with record type: \"%s\", record: \"%s\", template: \"%s\"",
		e.Type,
		e.URN,
		e.Template,
	)
}

type TemplateNotFoundError struct {
	URN string
}

func (e TemplateNotFoundError) Error() string {
	return fmt.Sprintf("could not find template \"%s\"", e.URN)
}

type DuplicateTaggingRecordError struct {
	RecordURN   string
	RecordType  string
	TemplateURN string
}

func (e DuplicateTaggingRecordError) Error() string {
	return fmt.Sprintf("tag of record URN \"%s\" with type \"%s\" and template URN \"%s\" already exists", e.RecordURN, e.RecordType, e.TemplateURN)
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
