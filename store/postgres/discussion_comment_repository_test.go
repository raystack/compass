package postgres_test

import (
	"github.com/odpf/columbus/discussion"
)

func (r *DiscussionRepositoryTestSuite) TestCreateComment() {
	r.Run("should create a new comment with full information", func() {
		cmt := &discussion.Comment{
			DiscussionID: "55555",
			Body:         "This is a new comment",
			Owner:        r.users[len(r.users)-1],
		}
		id, err := r.repository.CreateComment(r.ctx, cmt)
		r.NoError(err)
		r.NotEmpty(id)
	})

	r.Run("should return error when creating a new comment with empty body", func() {
		cmt := &discussion.Comment{
			Body:  "  ",
			Owner: r.users[len(r.users)-1],
		}
		id, err := r.repository.CreateComment(r.ctx, cmt)
		r.Error(err)
		r.Empty(id)
	})

	r.Run("should return error when creating a new comment with empty owner", func() {
		cmt := &discussion.Comment{
			Body: "This is Body",
		}
		id, err := r.repository.CreateComment(r.ctx, cmt)
		r.Error(err)
		r.Empty(id)
	})

	r.Run("should return error not found when creating a new comment with unknown discussion id", func() {
		cmt := &discussion.Comment{
			Body:  "some body",
			Owner: r.users[len(r.users)-1],
		}
		cmt.DiscussionID = "9278"
		id, err := r.repository.CreateComment(r.ctx, cmt)
		r.ErrorAs(err, new(discussion.NotFoundError))
		r.Empty(id)
	})
}

func (r *DiscussionRepositoryTestSuite) TestGetAllComments() {
	r.Run("should return list of comments of a discussion if discussion id exists", func() {
		discussionID := "11111"
		cmts, err := r.repository.GetAllComments(r.ctx, discussionID, discussion.Filter{})
		r.NoError(err)
		r.Len(cmts, 2)
		for _, cmt := range cmts {
			r.Equal(discussionID, cmt.DiscussionID)
		}
	})

	r.Run("should return empty list of comments of a discussion if discussion id does not exist", func() {
		discussionID := "90909"
		cmts, err := r.repository.GetAllComments(r.ctx, discussionID, discussion.Filter{})
		r.NoError(err)
		r.Empty(cmts)
	})

	r.Run("should return error if discussion id's type is wrong", func() {
		discussionID := "abc"
		cmts, err := r.repository.GetAllComments(r.ctx, discussionID, discussion.Filter{})
		r.Error(err)
		r.Empty(cmts)
	})

	r.Run("should working fine with filter", func() {
		testCases := []struct {
			description    string
			filter         discussion.Filter
			resultLength   int
			validateResult func(r *DiscussionRepositoryTestSuite, results []discussion.Comment)
		}{
			{
				description: "should limit with size",
				filter: discussion.Filter{
					Size: 1,
				},
				resultLength: 1,
			},
			{
				description: "should move cursor with offset",
				filter: discussion.Filter{
					Size:   5,
					Offset: 2,
				},
				resultLength: 1,
			},
			{
				description: "should sort descendingly with sort",
				filter: discussion.Filter{
					SortBy:        "updated_at",
					SortDirection: "desc",
				},
				resultLength: 3,
				validateResult: func(r *DiscussionRepositoryTestSuite, results []discussion.Comment) {
					r.Equal(results[0].ID, "55")
					r.Equal(results[1].ID, "44")
					r.Equal(results[2].ID, "33")
				},
			},
		}

		for _, testCase := range testCases {
			r.Run(testCase.description, func() {
				discussionID := "22222"
				dscs, err := r.repository.GetAllComments(r.ctx, discussionID, testCase.filter)
				r.NoError(err)
				r.Len(dscs, testCase.resultLength)

				if testCase.validateResult != nil {
					testCase.validateResult(r, dscs)
				}
			})
		}
	})
}

func (r *DiscussionRepositoryTestSuite) TestGetComment() {
	discussionID := "11111"
	r.Run("should return a comment if comment id exists", func() {
		commentID := "11"
		cmt, err := r.repository.GetComment(r.ctx, commentID, discussionID)
		r.NoError(err)
		r.Equal(commentID, cmt.ID)
		r.Equal(discussionID, cmt.DiscussionID)
		r.Equal("This is 1st comment of discussion 11111", cmt.Body)
	})

	r.Run("should return error not found if comment id does not exist", func() {
		commentID := "9090"
		cmt, err := r.repository.GetComment(r.ctx, commentID, discussionID)
		r.ErrorAs(err, new(discussion.NotFoundError))
		r.Empty(cmt)
	})

	r.Run("should return error if commnet id's type is wrong", func() {
		commentID := "abc"
		cmt, err := r.repository.GetComment(r.ctx, commentID, discussionID)
		r.Error(err)
		r.Empty(cmt)
	})
}

func (r *DiscussionRepositoryTestSuite) TestUpdateComment() {
	r.Run("should successfully update a comment", func() {
		cmt := &discussion.Comment{
			ID:           "55",
			DiscussionID: "22222",
			Body:         "Updated Body Comment",
			UpdatedBy:    r.users[1],
		}
		err := r.repository.UpdateComment(r.ctx, cmt)
		r.NoError(err)

		newCmt, err := r.repository.GetComment(r.ctx, cmt.ID, cmt.DiscussionID)
		r.NoError(err)
		r.Equal(newCmt.Body, cmt.Body)
		r.NotEqual(newCmt.UpdatedAt, cmt.UpdatedAt)
		r.Equal(newCmt.UpdatedBy.UUID, cmt.UpdatedBy.UUID)
	})

	r.Run("should return error when updating a comment that does not exist", func() {
		cmt := &discussion.Comment{
			ID:           "9090",
			DiscussionID: "22222",
			Body:         "Updated Body Comment",
			UpdatedBy:    r.users[len(r.users)-1],
		}
		err := r.repository.UpdateComment(r.ctx, cmt)
		r.Error(err)
	})

	r.Run("should return error when updating a comment that does not belong to a discussion", func() {
		cmt := &discussion.Comment{
			ID:           "55",
			DiscussionID: "9090",
			Body:         "Updated Body Comment",
			UpdatedBy:    r.users[len(r.users)-1],
		}
		err := r.repository.UpdateComment(r.ctx, cmt)
		r.Error(err)
	})

	r.Run("should return error when updating a comment with empty updated_by", func() {
		cmt := &discussion.Comment{
			Body: "This is Body",
		}
		err := r.repository.UpdateComment(r.ctx, cmt)
		r.Error(err)
	})
}

func (r *DiscussionRepositoryTestSuite) TestDeleteComment() {
	r.Run("should successfully delete a comment", func() {
		commentID := "55"
		discussionID := "22222"
		err := r.repository.DeleteComment(r.ctx, commentID, discussionID)
		r.NoError(err)

		newCmt, err := r.repository.GetComment(r.ctx, commentID, discussionID)
		r.ErrorAs(err, new(discussion.NotFoundError))
		r.Empty(newCmt)
	})

	r.Run("should return error when deleting a comment that does not exist", func() {
		commentID := "9090"
		discussionID := "22222"
		err := r.repository.DeleteComment(r.ctx, commentID, discussionID)
		r.Error(err)
	})

	r.Run("should return error when deleting a comment that does not belong to a discussion", func() {
		commentID := "55"
		discussionID := "9090"
		err := r.repository.DeleteComment(r.ctx, commentID, discussionID)
		r.Error(err)
	})
}
