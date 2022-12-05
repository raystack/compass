package asset_test

import (
	"testing"

	"github.com/odpf/compass/core/asset"
	"github.com/stretchr/testify/assert"
)

func TestValidateFilter(t *testing.T) {
	type testCase struct {
		Description string
		Filter      *asset.Filter
		errString   string
	}
	var testCases = []testCase{
		{
			Description: "empty filter will be valid",
			Filter:      &asset.Filter{},
		},
		{
			Description: "invalid type will return error",
			Filter:      &asset.Filter{Types: []asset.Type{"random"}},
			errString:   "error value \"random\" for key \"type\" not recognized, only support \"openended issues qanda all\"",
		},
		{
			Description: "invalid sort and direction will return error",
			Filter:      &asset.Filter{SortBy: "random", SortDirection: "random"},
			errString:   "error value \"random\" for key \"SortBy\" not recognized, only support \"name type service created_at updated_at\" and error value \"random\" for key \"SortDirection\" not recognized, only support \"asc desc\"",
		},
		{
			Description: "invalid size and offset will return error",
			Filter:      &asset.Filter{Size: -12, Offset: -1},
			errString:   "size cannot be less than 0 and offset cannot be less than 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			err := tc.Filter.Validate()
			if err != nil {
				assert.Equal(t, tc.errString, err.Error())
			}
		})
	}
}

func TestBuild(t *testing.T) {
	type testCase struct {
		Description   string
		ErrString     string
		Size          int
		SortBy        string
		SortDirection string
		Offset        int
		Types         string
		Services      string
		QFields       string
	}
	var testCases = []testCase{
		{
			Description: "invalid size and offset will return error",
			Size:        -12,
			Offset:      -1,
			ErrString:   "size cannot be less than 0 and offset cannot be less than 0",
		},
		{
			Description:   "invalid sort and direction will return error",
			SortBy:        "random",
			SortDirection: "random",
			ErrString:     "error value \"random\" for key \"SortBy\" not recognized, only support \"name type service created_at updated_at\" and error value \"random\" for key \"SortDirection\" not recognized, only support \"asc desc\"",
		},
		{
			Description: "returns no error for valid fields",
			Types:       "int,string,char",
			Services:    "dashboard",
			QFields:     "de-platform",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			fb := asset.NewFilterBuilder()
			fb.Data(map[string]string{
				"name": "go-data",
			})
			fb.Q("merchants")
			fb.Size(128)
			if tc.Offset != 0 {
				fb.Offset(tc.Offset)
			}
			if tc.QFields != "" {
				fb.QFields(tc.QFields)
			}
			if tc.Services != "" {
				fb.Services(tc.Services)
			}
			if tc.Size != 0 {
				fb.Size(tc.Size)
			}
			if tc.SortBy != "" {
				fb.SortBy(tc.SortBy)
			}
			if tc.SortDirection != "" {
				fb.SortDirection(tc.SortDirection)
			}
			if tc.Types != "" {
				fb.Types(tc.Types)
			}
			_, err := fb.Build()
			if err != nil {
				assert.Equal(t, tc.ErrString, err.Error())
			}
		})
	}
}
