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
					Name:           "three",
					Classification: models.TypeClassificationResource,
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
