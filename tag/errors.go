package tag

import (
	"fmt"
)

type ErrNotFound struct {
	URN      string
	Type     string
	Template string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf(
		"could not find tag with record type: \"%s\", record: \"%s\", template: \"%s\"",
		e.Type,
		e.URN,
		e.Template,
	)
}

type ErrTemplateNotFound struct {
	URN string
}

func (e ErrTemplateNotFound) Error() string {
	return fmt.Sprintf("could not find template \"%s\"", e.URN)
}

type ErrDuplicate struct {
	RecordURN   string
	RecordType  string
	TemplateURN string
}

func (e ErrDuplicate) Error() string {
	return fmt.Sprintf("tag of record URN \"%s\" with type \"%s\" and template URN \"%s\" already exists", e.RecordURN, e.RecordType, e.TemplateURN)
}

type ErrDuplicateTemplate struct {
	URN string
}

func (e ErrDuplicateTemplate) Error() string {
	return fmt.Sprintf("template \"%s\" already exists", e.URN)
}

type ErrValidation struct {
	Err error
}

func (e ErrValidation) Error() string {
	return e.Err.Error()
}
