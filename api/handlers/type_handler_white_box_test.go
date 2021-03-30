package handlers

import (
	"testing"

	"github.com/odpf/columbus/models"
)

func TestTypeHandler(t *testing.T) {
	t.Run("validateType", func(t *testing.T) {
		t.Run("should fail", func(t *testing.T) {
			types := []models.Type{
				{
					Name:           "one",
					Classification: models.TypeClassificationResource,
					Fields:         models.TypeFields{},
				},
				{
					Name:           "two",
					Classification: models.TypeClassificationResource,
					Fields: models.TypeFields{
						ID: "id",
					},
				},
				{
					Name:           "three",
					Classification: models.TypeClassificationResource,
					Fields: models.TypeFields{
						ID:    "id",
						Title: "name",
					},
					Lineage: []models.LineageDescriptor{
						{
							Type: "unknown",
							Dir:  models.DataflowDirUpstream,
						},
					},
				},
				{
					Name:           "four",
					Classification: models.TypeClassificationResource,
					Fields: models.TypeFields{
						ID:    "id",
						Title: "name",
					},
					Lineage: []models.LineageDescriptor{
						{
							Type:  "unknown",
							Query: "$.src",
						},
					},
				},
				{
					Name:           "four",
					Classification: models.TypeClassificationResource,
					Fields: models.TypeFields{
						ID:    "id",
						Title: "name",
					},
					Lineage: []models.LineageDescriptor{
						{
							Dir:   models.DataflowDirUpstream,
							Query: "$.src",
						},
					},
				},
			}

			handler := new(TypeHandler)
			for _, recordType := range types {
				err := handler.validateType(recordType)
				if err == nil {
					t.Errorf("expected type %#v to fail validation", recordType)
				}
			}
		})
	})
}
