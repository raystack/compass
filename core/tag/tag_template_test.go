package tag_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/tag"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTemplateToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		Template    tag.Template
		ExpectProto *compassv1beta1.TagTemplate
	}

	var testCases = []testCase{
		{
			Title:       "should return no timestamp pb and empty template field pb if timestamp and template field are empty",
			Template:    tag.Template{URN: "urn", DisplayName: "display-name", Description: "description"},
			ExpectProto: &compassv1beta1.TagTemplate{Urn: "urn", DisplayName: "display-name", Description: "description"},
		},
		{
			Title:       "should return timestamp pb and template field pb if timestamp and template field are not empty",
			Template:    tag.Template{URN: "urn", DisplayName: "display-name", Description: "description", Fields: []tag.Field{{ID: 12, URN: "urn1"}}, CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.TagTemplate{Urn: "urn", DisplayName: "display-name", Description: "description", Fields: []*compassv1beta1.TagTemplateField{{Id: 12, Urn: "urn1"}}, CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tc.Template.ToProto()
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestNewTemplateFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title  string
		PB     *compassv1beta1.TagTemplate
		Expect tag.Template
	}

	var testCases = []testCase{
		{
			Title:  "should return non empty time.Time and field if timestamp pb and field pb are not empty or zero",
			PB:     &compassv1beta1.TagTemplate{Urn: "urn", DisplayName: "display-name", Description: "description", Fields: []*compassv1beta1.TagTemplateField{{Id: 12, Urn: "urn1"}}, CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			Expect: tag.Template{URN: "urn", DisplayName: "display-name", Description: "description", Fields: []tag.Field{{ID: 12, URN: "urn1"}}, CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:  "should return empty time.Time and empty field if timestamp pb and field pb are empty or zero",
			PB:     &compassv1beta1.TagTemplate{Urn: "urn", DisplayName: "display-name", Description: "description"},
			Expect: tag.Template{URN: "urn", DisplayName: "display-name", Description: "description"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tag.NewTemplateFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}

func TestTemplateFieldToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		Field       tag.Field
		ExpectProto *compassv1beta1.TagTemplateField
	}

	var testCases = []testCase{
		{
			Title:       "should return no timestamp pb if timestamp is empty",
			Field:       tag.Field{ID: 123, URN: "urn"},
			ExpectProto: &compassv1beta1.TagTemplateField{Id: 123, Urn: "urn"},
		},
		{
			Title:       "should return timestamp pb if timestamp is not empty or zero",
			Field:       tag.Field{ID: 123, URN: "urn", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.TagTemplateField{Id: 123, Urn: "urn", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tc.Field.ToProto()
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestNewTemplateFieldFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title  string
		PB     *compassv1beta1.TagTemplateField
		Expect tag.Field
	}

	var testCases = []testCase{
		{
			Title:  "should return non empty time.Time if timestamp pb is not empty",
			PB:     &compassv1beta1.TagTemplateField{Id: 123, Urn: "urn", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			Expect: tag.Field{ID: 123, URN: "urn", CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:  "should return empty time.Time if timestamp pb is empty or zero",
			PB:     &compassv1beta1.TagTemplateField{Id: 123, Urn: "urn"},
			Expect: tag.Field{ID: 123, URN: "urn"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tag.NewTemplateFieldFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}
