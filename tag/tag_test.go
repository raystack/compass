package tag_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/tag"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTagToProto(t *testing.T) {
	type testCase struct {
		Title       string
		Tag         tag.Tag
		ExpectProto *compassv1beta1.Tag
	}

	var testCases = []testCase{
		{
			Title:       "should return empty field value pb if tag values is empty",
			Tag:         tag.Tag{RecordType: "type", RecordURN: "urn"},
			ExpectProto: &compassv1beta1.Tag{RecordType: "type", RecordUrn: "urn"},
		},
		{
			Title:       "should return tag value pb if tag values is not empty",
			Tag:         tag.Tag{RecordType: "type", RecordURN: "urn", TagValues: []tag.TagValue{{FieldID: 123, FieldURN: "urn"}}},
			ExpectProto: &compassv1beta1.Tag{RecordType: "type", RecordUrn: "urn", TagValues: []*compassv1beta1.TagValue{{FieldId: 123, FieldUrn: "urn"}}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got, err := tc.Tag.ToProto()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestNewTagFromProto(t *testing.T) {
	type testCase struct {
		Title  string
		PB     *compassv1beta1.Tag
		Expect tag.Tag
	}

	var testCases = []testCase{
		{
			Title:  "should return non empty tag values if tag values pb are not empty",
			PB:     &compassv1beta1.Tag{RecordType: "type", RecordUrn: "urn", TagValues: []*compassv1beta1.TagValue{{FieldId: 123, FieldUrn: "urn"}}},
			Expect: tag.Tag{RecordType: "type", RecordURN: "urn", TagValues: []tag.TagValue{{FieldID: 123, FieldURN: "urn"}}},
		},
		{
			Title:  "should return empty tag values if tag values pb are empty",
			PB:     &compassv1beta1.Tag{RecordType: "type", RecordUrn: "urn"},
			Expect: tag.Tag{RecordType: "type", RecordURN: "urn"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tag.NewFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}

func TestTagValueToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		TagValue    tag.TagValue
		ExpectProto *compassv1beta1.TagValue
	}

	var testCases = []testCase{
		{
			Title:       "should return no timestamp pb and empty field value pb if timestamp and field value are empty or zero",
			TagValue:    tag.TagValue{FieldID: 123, FieldURN: "urn"},
			ExpectProto: &compassv1beta1.TagValue{FieldId: 123, FieldUrn: "urn"},
		},
		{
			Title:       "should return timestamp pb and field value pb if timestamp and field value are not empty or zero",
			TagValue:    tag.TagValue{FieldID: 123, FieldURN: "urn", FieldValue: "a value", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.TagValue{FieldId: 123, FieldUrn: "urn", FieldValue: structpb.NewStringValue("a value"), CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got, err := tc.TagValue.ToProto()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestNewTagValueFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title  string
		PB     *compassv1beta1.TagValue
		Expect tag.TagValue
	}

	var testCases = []testCase{
		{
			Title:  "should return non empty time.Time and field value if timestamp pb and field value pb are not empty or zero",
			PB:     &compassv1beta1.TagValue{FieldId: 123, FieldUrn: "urn", FieldValue: structpb.NewStringValue("a value"), CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			Expect: tag.TagValue{FieldID: 123, FieldURN: "urn", FieldValue: "a value", CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:  "should return empty time.Time and empty field value if timestamp pb and field value pb are empty or zero",
			PB:     &compassv1beta1.TagValue{FieldId: 123, FieldUrn: "urn"},
			Expect: tag.TagValue{FieldID: 123, FieldURN: "urn"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tag.NewTagValueFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}
