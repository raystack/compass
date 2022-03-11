package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/comment"
	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type CommentRepositoryTestSuite struct {
	suite.Suite
	ctx            context.Context
	client         *postgres.Client
	pool           *dockertest.Pool
	resource       *dockertest.Resource
	repository     *postgres.CommentRepository
	discussionRepo *postgres.DiscussionRepository
	assetRepo      *postgres.AssetRepository
	userRepo       *postgres.UserRepository
	users          []user.User
	assets         []asset.Asset
}

func (r *CommentRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewLogrus()
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.userRepo, err = postgres.NewUserRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.assetRepo, err = postgres.NewAssetRepository(r.client, r.userRepo, defaultGetMaxSize, defaultProviderName)
	if err != nil {
		r.T().Fatal(err)
	}

	r.discussionRepo, err = postgres.NewDiscussionRepository(r.client, defaultGetMaxSize)
	if err != nil {
		r.T().Fatal(err)
	}

	r.repository, err = postgres.NewCommentRepository(r.client, defaultGetMaxSize)
	if err != nil {
		r.T().Fatal(err)
	}

	r.users, err = createUsers(r.userRepo, 5)
	if err != nil {
		r.T().Fatal(err)
	}

	r.assets, err = createAssets(r.assetRepo, r.users, asset.Type(asset.TypeTable.String()))
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *CommentRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	err := r.client.Close()
	if err != nil {
		r.T().Fatal(err)
	}
	err = purgeDocker(r.pool, r.resource)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *CommentRepositoryTestSuite) SetupTest() {
	if err := r.bootstrap(); err != nil {
		r.T().Fatal(err)
	}
}

func (r *CommentRepositoryTestSuite) TearDownTest() {
	if err := r.cleanup(); err != nil {
		r.T().Fatal(err)
	}
}

func (r *CommentRepositoryTestSuite) bootstrap() error {
	queries := []string{
		fmt.Sprintf(`insert into discussions values (11111, 'Kafka Source', 'We need to figure out how to source the new kafka', 1, 2, '%s', array['kafka','topic','question'],null,array['59bf4219-433e-4e38-ad72-ebab70c7ee7a'])`, r.users[0].ID),
		fmt.Sprintf(`insert into discussions values (22222, 'Missing data point on asset 1234-5678', 'Does anyone know why do we miss datapoint on asset 1234-5678', 1, 3, '%s', array['wondering','data','datapoint','question'],array['6663c3c7-62cf-41db-b0e1-4d443525f6d4'],array['59bf4219-433e-4e38-ad72-ebab70c7ee7a','b5b6b5fe-813f-45ac-92a2-d65b47377ea3'])`, r.users[1].ID),
		fmt.Sprintf(`insert into discussions values (33333, 'Improve data', 'How about cleaning the data more?', 1, 1, '%s', array['data','enhancement'],array['44a368e7-73df-4520-8803-3979a97f1cc3','59bf4219-433e-4e38-ad72-ebab70c7ee7a','458e8fdd-bca3-45a8-a17c-c00b2541d671'],array['ef1f0896-785a-47a0-8c9f-9897cdcf3697'])`, r.users[2].ID),
		fmt.Sprintf(`insert into discussions values (44444, 'Kafka Source (duplicated)', 'We need to figure out how to source the new kafka', 2, 2, '%s', array['kafka','topic'],null,array['ef1f0896-785a-47a0-8c9f-9897cdcf3697'])`, r.users[3].ID),
		fmt.Sprintf(`insert into discussions values (55555, 'Answered Questions', 'This question is answered', 2, 3, '%s', array['question','answered'],null,null)`, r.users[4].ID),
	}

	queries = append(queries, []string{
		fmt.Sprintf(`insert into comments values (11, 11111, 'This is 1st comment of discussion 11111', '%s', '%s')`, r.users[0].ID, r.users[0].ID),
		fmt.Sprintf(`insert into comments values (22, 11111, 'This is 2nd comment of discussion 11111', '%s', '%s')`, r.users[1].ID, r.users[1].ID),
		fmt.Sprintf(`insert into comments values (33, 22222, 'This is 1st comment of discussion 22222', '%s', '%s')`, r.users[2].ID, r.users[2].ID),
		fmt.Sprintf(`insert into comments values (44, 22222, 'This is 2nd comment of discussion 22222', '%s', '%s')`, r.users[3].ID, r.users[3].ID),
		fmt.Sprintf(`insert into comments values (55, 22222, 'This is 3rd comment of discussion 22222', '%s', '%s')`, r.users[0].ID, r.users[0].ID),
	}...)
	return r.client.ExecQueries(r.ctx, queries)
}

func (r *CommentRepositoryTestSuite) cleanup() error {
	queries := []string{
		"TRUNCATE TABLE discussions CASCADE",
	}
	return r.client.ExecQueries(r.ctx, queries)
}

func (r *CommentRepositoryTestSuite) TestCreate() {
	r.Run("should create a new comment with full information", func() {
		cmt := &comment.Comment{
			DiscussionID: "55555",
			Body:         "This is a new comment",
			Owner:        r.users[len(r.users)-1],
		}
		id, err := r.repository.Create(r.ctx, cmt)
		r.NoError(err)
		r.NotEmpty(id)
	})

	r.Run("should return error when creating a new comment with empty body", func() {
		cmt := &comment.Comment{
			Body:  "  ",
			Owner: r.users[len(r.users)-1],
		}
		id, err := r.repository.Create(r.ctx, cmt)
		r.Error(err)
		r.Empty(id)
	})

	r.Run("should return error when creating a new comment with empty owner", func() {
		cmt := &comment.Comment{
			Body: "This is Body",
		}
		id, err := r.repository.Create(r.ctx, cmt)
		r.Error(err)
		r.Empty(id)
	})
}

func (r *CommentRepositoryTestSuite) TestGetAll() {
	r.Run("should return list of comments of a discussion if discussion id exists", func() {
		discussionID := "11111"
		cmts, err := r.repository.GetAll(r.ctx, discussionID, comment.Filter{})
		r.NoError(err)
		r.Len(cmts, 2)
		for _, cmt := range cmts {
			r.Equal(discussionID, cmt.DiscussionID)
		}
	})

	r.Run("should return empty list of comments of a discussion if discussion id does not exist", func() {
		discussionID := "90909"
		cmts, err := r.repository.GetAll(r.ctx, discussionID, comment.Filter{})
		r.NoError(err)
		r.Empty(cmts)
	})

	r.Run("should return error if discussion id's type is wrong", func() {
		discussionID := "abc"
		cmts, err := r.repository.GetAll(r.ctx, discussionID, comment.Filter{})
		r.Error(err)
		r.Empty(cmts)
	})

	r.Run("should working fine with filter", func() {
		testCases := []struct {
			description    string
			filter         comment.Filter
			resultLength   int
			validateResult func(r *CommentRepositoryTestSuite, results []comment.Comment)
		}{
			{
				description: "should limit with size",
				filter: comment.Filter{
					Size: 1,
				},
				resultLength: 1,
			},
			{
				description: "should move cursor with offset",
				filter: comment.Filter{
					Size:   5,
					Offset: 2,
				},
				resultLength: 1,
			},
			{
				description: "should sort descendingly with sort",
				filter: comment.Filter{
					SortBy:        "updated_at",
					SortDirection: "desc",
				},
				resultLength: 3,
				validateResult: func(r *CommentRepositoryTestSuite, results []comment.Comment) {
					r.Equal(results[0].ID, "55")
					r.Equal(results[1].ID, "44")
					r.Equal(results[2].ID, "33")
				},
			},
		}

		for _, testCase := range testCases {
			r.Run(testCase.description, func() {
				discussionID := "22222"
				dscs, err := r.repository.GetAll(r.ctx, discussionID, testCase.filter)
				r.NoError(err)
				r.Len(dscs, testCase.resultLength)

				if testCase.validateResult != nil {
					testCase.validateResult(r, dscs)
				}
			})
		}
	})
}

func (r *CommentRepositoryTestSuite) TestGet() {
	discussionID := "11111"
	r.Run("should return a comment if comment id exists", func() {
		commentID := "11"
		cmt, err := r.repository.Get(r.ctx, commentID, discussionID)
		r.NoError(err)
		r.Equal(commentID, cmt.ID)
		r.Equal(discussionID, cmt.DiscussionID)
		r.Equal("This is 1st comment of discussion 11111", cmt.Body)
	})

	r.Run("should return error not found if comment id does not exist", func() {
		commentID := "9090"
		cmt, err := r.repository.Get(r.ctx, commentID, discussionID)
		r.ErrorAs(err, new(comment.NotFoundError))
		r.Empty(cmt)
	})

	r.Run("should return error if commnet id's type is wrong", func() {
		commentID := "abc"
		cmt, err := r.repository.Get(r.ctx, commentID, discussionID)
		r.Error(err)
		r.Empty(cmt)
	})
}

func (r *CommentRepositoryTestSuite) TestUpdate() {
	r.Run("should successfully update a comment", func() {
		cmt := &comment.Comment{
			ID:           "55",
			DiscussionID: "22222",
			Body:         "Updated Body Comment",
			UpdatedBy:    r.users[1],
		}
		err := r.repository.Update(r.ctx, cmt)
		r.NoError(err)

		newCmt, err := r.repository.Get(r.ctx, cmt.ID, cmt.DiscussionID)
		r.NoError(err)
		r.Equal(newCmt.Body, cmt.Body)
		r.NotEqual(newCmt.UpdatedAt, cmt.UpdatedAt)
		r.Equal(newCmt.UpdatedBy.ID, cmt.UpdatedBy.ID)
	})

	r.Run("should return error when updating a comment that does not exist", func() {
		cmt := &comment.Comment{
			ID:           "9090",
			DiscussionID: "22222",
			Body:         "Updated Body Comment",
			UpdatedBy:    r.users[len(r.users)-1],
		}
		err := r.repository.Update(r.ctx, cmt)
		r.Error(err)
	})

	r.Run("should return error when updating a comment that does not belong to a discussion", func() {
		cmt := &comment.Comment{
			ID:           "55",
			DiscussionID: "9090",
			Body:         "Updated Body Comment",
			UpdatedBy:    r.users[len(r.users)-1],
		}
		err := r.repository.Update(r.ctx, cmt)
		r.Error(err)
	})

	r.Run("should return error when updating a comment with empty updated_by", func() {
		cmt := &comment.Comment{
			Body: "This is Body",
		}
		err := r.repository.Update(r.ctx, cmt)
		r.Error(err)
	})
}

func (r *CommentRepositoryTestSuite) TestDelete() {
	r.Run("should successfully delete a comment", func() {
		commentID := "55"
		discussionID := "22222"
		err := r.repository.Delete(r.ctx, commentID, discussionID)
		r.NoError(err)

		newCmt, err := r.repository.Get(r.ctx, commentID, discussionID)
		r.ErrorAs(err, new(comment.NotFoundError))
		r.Empty(newCmt)
	})

	r.Run("should return error when deleting a comment that does not exist", func() {
		commentID := "9090"
		discussionID := "22222"
		err := r.repository.Delete(r.ctx, commentID, discussionID)
		r.Error(err)
	})

	r.Run("should return error when deleting a comment that does not belong to a discussion", func() {
		commentID := "55"
		discussionID := "9090"
		err := r.repository.Delete(r.ctx, commentID, discussionID)
		r.Error(err)
	})
}

func TestCommentRepository(t *testing.T) {
	suite.Run(t, &CommentRepositoryTestSuite{})
}
