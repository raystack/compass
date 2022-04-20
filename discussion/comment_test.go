package discussion_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/discussion"
	"github.com/odpf/compass/user"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCommentToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		Comment     *discussion.Comment
		ExpectProto *compassv1beta1.Comment
	}

	var testCases = []testCase{
		{
			Title:       "should return no timestamp pb if timestamp is zero",
			Comment:     &discussion.Comment{ID: "id1"},
			ExpectProto: &compassv1beta1.Comment{Id: "id1"},
		},
		{
			Title:       "should return timestamp pb if timestamp is not zero",
			Comment:     &discussion.Comment{ID: "id1", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.Comment{Id: "id1", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tc.Comment.ToProto()
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestNewCommentFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title  string
		PB     *compassv1beta1.Comment
		Expect discussion.Comment
	}

	var testCases = []testCase{
		{
			Title:  "should return non empty time.Time, owner, and updated by if pb is not empty or zero",
			PB:     &compassv1beta1.Comment{Id: "id1", Owner: &compassv1beta1.User{Id: "uid1"}, UpdatedBy: &compassv1beta1.User{Id: "uid1"}, CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			Expect: discussion.Comment{ID: "id1", Owner: user.User{ID: "uid1"}, UpdatedBy: user.User{ID: "uid1"}, CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:  "should return empty time.Time, owner, and updated by if pb is empty or zero",
			PB:     &compassv1beta1.Comment{Id: "id1"},
			Expect: discussion.Comment{ID: "id1"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := discussion.NewCommentFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}
