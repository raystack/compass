package discussion_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/user"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestIsEmpty(t *testing.T) {
	type TestCase struct {
		Description string
		Discussion  discussion.Discussion
		IsEmpty     bool
	}

	var testCases = []TestCase{
		{
			Description: "all necessary fields are empty and nil will be considered empty",
			Discussion:  discussion.Discussion{ID: "123", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			IsEmpty:     true,
		},
		{
			Description: "nil slice will be considered empty",
			Discussion:  discussion.Discussion{Labels: nil},
			IsEmpty:     true,
		},
		{
			Description: "empty slice won't be considered empty",
			Discussion:  discussion.Discussion{Labels: []string{}},
			IsEmpty:     false,
		},
		{
			Description: "title exist won't be considered empty",
			Discussion:  discussion.Discussion{Title: "title"},
			IsEmpty:     false,
		},
		{
			Description: "body exist won't be considered empty",
			Discussion:  discussion.Discussion{Body: "body"},
			IsEmpty:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			assert.Equal(t, tc.IsEmpty, tc.Discussion.IsEmpty())
		})
	}
}

func TestValidateConstraint(t *testing.T) {
	type TestCase struct {
		Description string
		Discussion  discussion.Discussion
		Err         error
	}

	var testCases = []TestCase{
		{
			Description: "type is not one of supported types will return error",
			Discussion:  discussion.Discussion{Type: "random"},
			Err:         discussion.ErrInvalidType,
		},
		{
			Description: "state is not one of supported states will return error",
			Discussion:  discussion.Discussion{State: "random"},
			Err:         discussion.ErrInvalidState,
		},
		{
			Description: "labels is more than MAX will return error",
			Discussion:  discussion.Discussion{Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}},
			Err:         errors.New("labels cannot be more than 10"),
		},
		{
			Description: "assets is more than MAX will return error",
			Discussion:  discussion.Discussion{Assets: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}},
			Err:         errors.New("assets cannot be more than 10"),
		},
		{
			Description: "assignees is more than MAX will return error",
			Discussion:  discussion.Discussion{Assignees: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}},
			Err:         errors.New("assignees cannot be more than 10"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			assert.Equal(t, tc.Err, tc.Discussion.ValidateConstraint())
		})
	}
}

func TestValidateDiscussion(t *testing.T) {
	type TestCase struct {
		Description string
		Discussion  discussion.Discussion
		Err         error
	}

	var testCases = []TestCase{
		{
			Description: "empty title will return error",
			Discussion:  discussion.Discussion{},
			Err:         errors.New("title cannot be empty"),
		},
		{
			Description: "empty body will return error",
			Discussion:  discussion.Discussion{Title: "title"},
			Err:         errors.New("body cannot be empty"),
		},
		{
			Description: "empty type will return error",
			Discussion:  discussion.Discussion{Title: "title", Body: "body"},
			Err:         errors.New("type must be specified"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			assert.Equal(t, tc.Err, tc.Discussion.Validate())
		})
	}
}

func TestToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		Discussion  *discussion.Discussion
		ExpectProto *compassv1beta1.Discussion
	}

	var testCases = []testCase{
		{
			Title:       "should return no timestamp pb if timestamp is zero",
			Discussion:  &discussion.Discussion{ID: "id1"},
			ExpectProto: &compassv1beta1.Discussion{Id: "id1"},
		},
		{
			Title:       "should return timestamp pb if timestamp is not zero",
			Discussion:  &discussion.Discussion{ID: "id1", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.Discussion{Id: "id1", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tc.Discussion.ToProto()
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestNewFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title            string
		DiscussionPB     *compassv1beta1.Discussion
		ExpectDiscussion discussion.Discussion
	}

	var testCases = []testCase{
		{
			Title:            "should return non empty time.Time and owner if pb is not empty or zero",
			DiscussionPB:     &compassv1beta1.Discussion{Id: "id1", Owner: &compassv1beta1.User{Id: "uid1"}, CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			ExpectDiscussion: discussion.Discussion{ID: "id1", Owner: user.User{ID: "uid1"}, Type: "openended", State: "open", CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:            "should return empty time.Time and owner if pb is empty or zero",
			DiscussionPB:     &compassv1beta1.Discussion{Id: "id1"},
			ExpectDiscussion: discussion.Discussion{ID: "id1", Type: "openended", State: "open"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := discussion.NewFromProto(tc.DiscussionPB)
			if reflect.DeepEqual(got, tc.ExpectDiscussion) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.ExpectDiscussion, got)
			}
		})
	}
}
