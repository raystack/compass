package asset

import (
	"strings"

	"github.com/odpf/compass/core/validator"
)

type Filter struct {
	Types         []Type
	Services      []string
	Size          int
	Offset        int
	SortBy        string `validate:"omitempty,oneof=name type service created_at updated_at"`
	SortDirection string `validate:"omitempty,oneof=asc desc"`
	QueryFields   []string
	Query         string
	Data          map[string][]string
}

func (f *Filter) Validate() error {
	return validator.ValidateStruct(f)
}

type filterBuilder struct {
	types         string
	services      string
	q             string
	qFields       string
	data          map[string]string
	size          int
	offset        int
	sortBy        string
	sortDirection string
}

func NewFilterBuilder() *filterBuilder {
	return &filterBuilder{}
}

func (fb *filterBuilder) Types(types string) *filterBuilder {
	fb.types = types
	return fb
}

func (fb *filterBuilder) Services(services string) *filterBuilder {
	fb.services = services
	return fb
}

func (fb *filterBuilder) Q(q string) *filterBuilder {
	fb.q = q
	return fb
}

func (fb *filterBuilder) QFields(qFields string) *filterBuilder {
	fb.qFields = qFields
	return fb
}

func (fb *filterBuilder) Data(data map[string]string) *filterBuilder {
	fb.data = data
	return fb
}

func (fb *filterBuilder) Size(size int) *filterBuilder {
	fb.size = size
	return fb
}

func (fb *filterBuilder) Offset(offset int) *filterBuilder {
	fb.offset = offset
	return fb
}

func (fb *filterBuilder) SortBy(sortBy string) *filterBuilder {
	fb.sortBy = sortBy
	return fb
}

func (fb *filterBuilder) SortDirection(sortDirection string) *filterBuilder {
	fb.sortDirection = sortDirection
	return fb
}

func (fb *filterBuilder) Build() (Filter, error) {
	flt := Filter{
		Size:          fb.size,
		Offset:        fb.offset,
		SortBy:        fb.sortBy,
		SortDirection: fb.sortDirection,
		Query:         fb.q,
	}

	if len(fb.data) != 0 {
		flt.Data = make(map[string][]string)
		for k, v := range fb.data {
			flt.Data[k] = strings.Split(v, ",")
		}
	}

	if fb.types != "" {
		typs := strings.Split(fb.types, ",")
		for _, typeVal := range typs {
			flt.Types = append(flt.Types, Type(typeVal))
		}
	}
	if fb.services != "" {
		flt.Services = strings.Split(fb.services, ",")
	}

	if fb.qFields != "" {
		flt.QueryFields = strings.Split(fb.qFields, ",")
	}

	if err := flt.Validate(); err != nil {
		return Filter{}, err
	}

	return flt, nil
}
