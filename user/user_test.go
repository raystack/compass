package user

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidate(t *testing.T) {
	type testCase struct {
		Title       string
		User        *User
		ExpectError error
	}

	var testCases = []testCase{
		{
			Title:       "should return error no user information if user is nil",
			User:        nil,
			ExpectError: ErrNoUserInformation,
		},
		{
			Title:       "should return error invalid if uuid is empty",
			User:        &User{Provider: "provider"},
			ExpectError: InvalidError{},
		},
		{
			Title:       "should return nil if user is valid",
			User:        &User{UUID: "some-uuid", Provider: "provider"},
			ExpectError: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {

			err := testCase.User.Validate()
			assert.Equal(t, testCase.ExpectError, err)
		})
	}
}

func TestToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		User        *User
		ExpectProto *compassv1beta1.User
	}

	var testCases = []testCase{
		{
			Title:       "should return nil if UUID is empty",
			User:        &User{},
			ExpectProto: nil,
		},
		{
			Title:       "should return fields without timestamp",
			User:        &User{UUID: "uuid1", Email: "email@email.com", Provider: "provider", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.User{Uuid: "uuid1", Email: "email@email.com"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tc.User.ToProto()
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestToFullProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		User        *User
		ExpectProto *compassv1beta1.User
	}

	var testCases = []testCase{
		{
			Title:       "should return nil if UUID is empty",
			User:        &User{},
			ExpectProto: nil,
		},
		{
			Title:       "should return without timestamp pb if timestamp is zero",
			User:        &User{UUID: "uuid1", Provider: "provider"},
			ExpectProto: &compassv1beta1.User{Uuid: "uuid1", Provider: "provider"},
		},
		{
			Title:       "should return with timestamp pb if timestamp is not zero",
			User:        &User{UUID: "uuid1", Provider: "provider", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.User{Uuid: "uuid1", Provider: "provider", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tc.User.ToFullProto()
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestNewFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title      string
		UserPB     *compassv1beta1.User
		ExpectUser User
	}

	var testCases = []testCase{
		{
			Title:      "should return non empty time.Time if timestamp pb is not zero",
			UserPB:     &compassv1beta1.User{Uuid: "uuid1", Provider: "provider", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			ExpectUser: User{UUID: "uuid1", Provider: "provider", CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:      "should return empty time.Time if timestamp pb is zero",
			UserPB:     &compassv1beta1.User{Uuid: "uuid1", Provider: "provider"},
			ExpectUser: User{UUID: "uuid1", Provider: "provider"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := NewFromProto(tc.UserPB)
			if reflect.DeepEqual(got, tc.ExpectUser) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.ExpectUser, got)
			}
		})
	}
}
