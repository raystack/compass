package discussion_test

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/compass/core/discussion"
	"github.com/odpf/compass/core/discussion/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetDiscussions(t *testing.T) {
	ctx := context.Background()
	testcase := []struct {
		Description string
		Setup       func(context.Context, *mocks.DiscussionRepository)
		Err         string
		Filter      discussion.Filter
		Discussion  []discussion.Discussion
	}{
		{
			Description: "should catch not found error",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{}).Return([]discussion.Discussion{}, discussion.NotFoundError{})
			},
			Err:        discussion.NotFoundError{}.Error(),
			Discussion: []discussion.Discussion{},
			Filter:     discussion.Filter{},
		},
		{
			Description: "should catch invalid error for invalid state",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{Type: "random"}).Return([]discussion.Discussion{}, discussion.InvalidError{})
			},
			Err:        discussion.InvalidError{}.Error(),
			Discussion: []discussion.Discussion{},
			Filter:     discussion.Filter{Type: "random"},
		},
		{
			Description: "should return all discussions for correct request",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{}).Return([]discussion.Discussion{{ID: "1", Title: "title", Body: "body"}}, nil)
			},
			Filter: discussion.Filter{},
			Discussion: []discussion.Discussion{
				{ID: "1", Title: "title", Body: "body"},
			},
		},
	}

	for _, tc := range testcase {
		t.Run(tc.Description, func(t *testing.T) {
			mockDiscussionRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscussionRepo)
			}
			defer mockDiscussionRepo.AssertExpectations(t)

			svc := discussion.NewService(mockDiscussionRepo)
			discussions, err := svc.GetDiscussions(ctx, tc.Filter)
			if err != nil {
				assert.Equal(t, tc.Err, err.Error())
			}
			assert.Equal(t, tc.Discussion, discussions)
		})
	}
}

func TestGetDiscussion(t *testing.T) {
	ctx := context.Background()
	testcase := []struct {
		Description string
		Setup       func(context.Context, *mocks.DiscussionRepository)
		Err         string
		did         string
		Discussion  discussion.Discussion
	}{
		{
			Description: "should catch not found error",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Get(ctx, "id-1").Return(discussion.Discussion{}, discussion.NotFoundError{DiscussionID: "id-1"})
			},
			Err:        discussion.NotFoundError{DiscussionID: "id-1"}.Error(),
			Discussion: discussion.Discussion{},
			did:        "id-1",
		},
		{
			Description: "should catch invalid error for invalid state",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Get(ctx, "invalid-id").Return(discussion.Discussion{}, discussion.InvalidError{DiscussionID: "invalid-id"})
			},
			Err:        discussion.InvalidError{DiscussionID: "invalid-id"}.Error(),
			Discussion: discussion.Discussion{},
			did:        "invalid-id",
		},
		{
			Description: "should return all discussions for correct request",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Get(ctx, "id").Return(discussion.Discussion{ID: "1", Title: "title", Body: "body"}, nil)
			},
			Discussion: discussion.Discussion{
				ID: "1", Title: "title", Body: "body",
			},
			did: "id",
		},
	}

	for _, tc := range testcase {
		t.Run(tc.Description, func(t *testing.T) {
			mockDiscussionRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscussionRepo)
			}
			defer mockDiscussionRepo.AssertExpectations(t)

			svc := discussion.NewService(mockDiscussionRepo)
			discussions, err := svc.GetDiscussion(ctx, tc.did)
			if err != nil {
				assert.Equal(t, tc.Err, err.Error())
			}
			assert.Equal(t, tc.Discussion, discussions)
		})
	}
}

func TestGetComments(t *testing.T) {
	ctx := context.Background()
	testcase := []struct {
		Description string
		Setup       func(context.Context, *mocks.DiscussionRepository)
		Err         string
		Filter      discussion.Filter
		Comment     []discussion.Comment
	}{
		{
			Description: "should catch not found error",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAllComments(ctx, "id-1", discussion.Filter{}).Return([]discussion.Comment{}, discussion.NotFoundError{})
			},
			Err:     discussion.NotFoundError{}.Error(),
			Comment: []discussion.Comment{},
			Filter:  discussion.Filter{},
		},
		{
			Description: "should catch invalid error for invalid state",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAllComments(ctx, "id-1", discussion.Filter{Type: "random"}).Return([]discussion.Comment{}, discussion.InvalidError{})
			},
			Err:     discussion.InvalidError{}.Error(),
			Comment: []discussion.Comment{},
			Filter:  discussion.Filter{Type: "random"},
		},
		{
			Description: "should return all discussions for correct request",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAllComments(ctx, "id-1", discussion.Filter{}).Return([]discussion.Comment{{ID: "1", DiscussionID: "id-1", Body: "body"}}, nil)
			},
			Comment: []discussion.Comment{
				{ID: "1", DiscussionID: "id-1", Body: "body"},
			},
			Filter: discussion.Filter{},
		},
	}

	for _, tc := range testcase {
		t.Run(tc.Description, func(t *testing.T) {
			mockDiscussionRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscussionRepo)
			}
			defer mockDiscussionRepo.AssertExpectations(t)

			svc := discussion.NewService(mockDiscussionRepo)
			comments, err := svc.GetComments(ctx, "id-1", tc.Filter)
			if err != nil {
				assert.Equal(t, tc.Err, err.Error())
			}
			assert.Equal(t, tc.Comment, comments)
		})
	}
}

func TestGetDeleteComment(t *testing.T) {
	ctx := context.Background()
	testcase := []struct {
		Description string
		Setup       func(context.Context, *mocks.DiscussionRepository)
		Err         string
		cid         string
		did         string
		Comment     discussion.Comment
	}{
		{
			Description: "should catch not found error",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetComment(ctx, "id-1", "id-1").Return(discussion.Comment{}, discussion.NotFoundError{DiscussionID: "id-1", CommentID: "id-1"})
			},
			Err:     discussion.NotFoundError{DiscussionID: "id-1", CommentID: "id-1"}.Error(),
			Comment: discussion.Comment{},
			cid:     "id-1",
			did:     "id-1",
		},
		{
			Description: "should catch invalid error for invalid state",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetComment(ctx, "invalid-id", "invalid-id").Return(discussion.Comment{}, discussion.InvalidError{DiscussionID: "invalid-id", CommentID: "invalid-id"})
			},
			Err:     discussion.InvalidError{DiscussionID: "invalid-id", CommentID: "invalid-id"}.Error(),
			Comment: discussion.Comment{},
			cid:     "invalid-id",
			did:     "invalid-id",
		},
		{
			Description: "should return all discussions for correct request",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetComment(ctx, "id", "id").Return(discussion.Comment{ID: "1", DiscussionID: "id", Body: "body"}, nil)
				dr.EXPECT().DeleteComment(ctx, "id", "id").Return(nil)
			},
			Comment: discussion.Comment{
				ID: "1", DiscussionID: "id", Body: "body",
			},
			cid: "id",
			did: "id",
		},
	}

	for _, tc := range testcase {
		t.Run(tc.Description, func(t *testing.T) {
			mockDiscussionRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscussionRepo)
			}
			defer mockDiscussionRepo.AssertExpectations(t)

			svc := discussion.NewService(mockDiscussionRepo)
			comment, err := svc.GetComment(ctx, tc.did, tc.cid)
			if err != nil {
				assert.Equal(t, tc.Err, err.Error())
			} else {
				err = svc.DeleteComment(ctx, tc.did, tc.cid)
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.Comment, comment)
		})
	}
}

func TestCreateDiscussions(t *testing.T) {
	ctx := context.Background()
	validDiscussion := discussion.Discussion{
		ID: "1", Title: "title", Body: "body", Type: "openended",
	}
	invalidDiscussion := discussion.Discussion{
		ID: "1", Title: "title", Body: "body", Type: "invalid",
	}
	testcase := []struct {
		Description string
		Setup       func(context.Context, *mocks.DiscussionRepository)
		Err         error
		Filter      discussion.Filter
		Discussion  discussion.Discussion
	}{
		{
			Description: "throws error for empty discussion",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Create(ctx, &discussion.Discussion{}).Return("", errors.New("empty fields"))
			},
			Err:        errors.New("empty fields"),
			Discussion: discussion.Discussion{},
		},
		{
			Description: "throws error for invalid type",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Create(ctx, &invalidDiscussion).Return("", discussion.InvalidError{})
			},
			Err:        discussion.InvalidError{},
			Discussion: invalidDiscussion,
		},
		{
			Description: "create discussion without error for correct input",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Create(ctx, &validDiscussion).Return("", nil)
			},
			Discussion: validDiscussion,
		},
	}

	for _, tc := range testcase {
		t.Run(tc.Description, func(t *testing.T) {
			mockDiscussionRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscussionRepo)
			}
			defer mockDiscussionRepo.AssertExpectations(t)

			svc := discussion.NewService(mockDiscussionRepo)
			_, err := svc.CreateDiscussion(ctx, &tc.Discussion)
			assert.Equal(t, tc.Err, err)
		})
	}
}

func TestPatchDiscussions(t *testing.T) {
	ctx := context.Background()
	validDiscussion := discussion.Discussion{
		ID: "1", Title: "title", Body: "body", Type: "openended",
	}
	invalidDiscussion := discussion.Discussion{
		ID: "1", Title: "title", Body: "body", Type: "invalid",
	}
	testcase := []struct {
		Description string
		Setup       func(context.Context, *mocks.DiscussionRepository)
		Err         error
		Filter      discussion.Filter
		Discussion  discussion.Discussion
	}{
		{
			Description: "throws error for empty discussion",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Patch(ctx, &discussion.Discussion{}).Return(errors.New("empty fields"))
			},
			Err:        errors.New("empty fields"),
			Discussion: discussion.Discussion{},
		},
		{
			Description: "throws error for invalid type",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Patch(ctx, &invalidDiscussion).Return(discussion.InvalidError{})
			},
			Err:        discussion.InvalidError{},
			Discussion: invalidDiscussion,
		},
		{
			Description: "patch discussion properly for valid data",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Patch(ctx, &validDiscussion).Return(nil)
			},
			Discussion: validDiscussion,
		},
	}

	for _, tc := range testcase {
		t.Run(tc.Description, func(t *testing.T) {
			mockDiscussionRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscussionRepo)
			}
			defer mockDiscussionRepo.AssertExpectations(t)

			svc := discussion.NewService(mockDiscussionRepo)
			err := svc.PatchDiscussion(ctx, &tc.Discussion)
			assert.Equal(t, tc.Err, err)
		})
	}
}

func TestCreateComment(t *testing.T) {
	ctx := context.Background()
	validComment := discussion.Comment{
		ID: "1", DiscussionID: "id-1", Body: "body",
	}
	testcase := []struct {
		Description string
		Setup       func(context.Context, *mocks.DiscussionRepository)
		Err         error
		Comment     discussion.Comment
	}{
		{
			Description: "throw error for empty fields",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().CreateComment(ctx, &discussion.Comment{}).Return("", errors.New("empty fields"))
			},
			Err:     errors.New("empty fields"),
			Comment: discussion.Comment{},
		},
		{
			Description: "create comment for proper data",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().CreateComment(ctx, &validComment).Return("", nil)
			},
			Comment: validComment,
		},
	}

	for _, tc := range testcase {
		t.Run(tc.Description, func(t *testing.T) {
			mockDiscussionRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscussionRepo)
			}
			defer mockDiscussionRepo.AssertExpectations(t)

			svc := discussion.NewService(mockDiscussionRepo)
			_, err := svc.CreateComment(ctx, &tc.Comment)
			assert.Equal(t, tc.Err, err)
		})
	}
}

func TestUpdateComment(t *testing.T) {
	ctx := context.Background()
	validComment := discussion.Comment{
		ID: "1", DiscussionID: "id-1", Body: "body",
	}
	testcase := []struct {
		Description string
		Setup       func(context.Context, *mocks.DiscussionRepository)
		Err         error
		Comment     discussion.Comment
	}{
		{
			Description: "throw error for empty fields",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().UpdateComment(ctx, &discussion.Comment{}).Return(errors.New("empty fields"))
			},
			Err:     errors.New("empty fields"),
			Comment: discussion.Comment{},
		},
		{
			Description: "update comment for proper data",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().UpdateComment(ctx, &validComment).Return(nil)
			},
			Comment: validComment,
		},
	}

	for _, tc := range testcase {
		t.Run(tc.Description, func(t *testing.T) {
			mockDiscussionRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscussionRepo)
			}
			defer mockDiscussionRepo.AssertExpectations(t)

			svc := discussion.NewService(mockDiscussionRepo)
			err := svc.UpdateComment(ctx, &tc.Comment)
			assert.Equal(t, tc.Err, err)
		})
	}
}
