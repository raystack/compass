package models_test

import (
	"testing"

	"github.com/odpf/columbus/models"
)

func TestDataflowDir(t *testing.T) {
	var validStrGenerator = func(v bool) string {
		if v == true {
			return "valid"
		}
		return "invalid"
	}
	t.Run("test valid values", func(t *testing.T) {
		var testCases = []struct {
			Dir   string
			Valid bool
		}{
			{
				Dir:   "upstream",
				Valid: true,
			},
			{
				Dir:   "downstream",
				Valid: true,
			},
			{
				Dir: "bidirectional",
			},
			{
				Dir: "up",
			},
			{
				Dir: "down",
			},
		}

		for _, testCase := range testCases {
			value := models.DataflowDir(testCase.Dir)
			if value.Valid() != testCase.Valid {
				t.Errorf(
					"expected DataflowDir(%q) to be %s, was %s",
					testCase.Dir,
					validStrGenerator(testCase.Valid),
					validStrGenerator(value.Valid()),
				)
			}
		}
	})
}
