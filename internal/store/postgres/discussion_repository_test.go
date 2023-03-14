package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/discussion"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type DiscussionRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.DiscussionRepository
	assetRepo  *postgres.AssetRepository
	userRepo   *postgres.UserRepository
	users      []user.User
	assets     []asset.Asset
}

func (r *DiscussionRepositoryTestSuite) SetupSuite() {
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

	r.repository, err = postgres.NewDiscussionRepository(r.client, defaultGetMaxSize)
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

func (r *DiscussionRepositoryTestSuite) TearDownSuite() {
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

func (r *DiscussionRepositoryTestSuite) SetupTest() {
	if err := r.bootstrap(); err != nil {
		r.T().Fatal(err)
	}
}

func (r *DiscussionRepositoryTestSuite) TearDownTest() {
	if err := r.cleanup(); err != nil {
		r.T().Fatal(err)
	}
}

func (r *DiscussionRepositoryTestSuite) bootstrap() error {
	queries := []string{
		fmt.Sprintf(`insert into discussions values (11111, 'Kafka Source', 'We need to figure out how to source the new kafka', 1, 2, '%s', array['kafka','topic','question'],null,array['59bf4219-433e-4e38-ad72-ebab70c7ee7a'])`, r.users[0].ID),
		fmt.Sprintf(`insert into discussions values (22222, 'Missing data point on asset 1234-5678', 'Does anyone know why do we miss datapoint on asset 1234-5678', 'open', 'qanda', '%s', array['wondering','data','datapoint','question'],array['6663c3c7-62cf-41db-b0e1-4d443525f6d4'],array['59bf4219-433e-4e38-ad72-ebab70c7ee7a','%s'])`, r.users[1].ID, r.users[0].ID),
		fmt.Sprintf(`insert into discussions values (33333, 'Improve data', 'How about cleaning the data more?', 'open', 'openended', '%s', array['data','enhancement'],array['44a368e7-73df-4520-8803-3979a97f1cc3','59bf4219-433e-4e38-ad72-ebab70c7ee7a','458e8fdd-bca3-45a8-a17c-c00b2541d671'],array['ef1f0896-785a-47a0-8c9f-9897cdcf3697'])`, r.users[2].ID),
		fmt.Sprintf(`insert into discussions values (44444, 'Kafka Source (duplicated)', 'We need to figure out how to source the new kafka', 'closed', 'issues', '%s', array['kafka','topic'],null,array['ef1f0896-785a-47a0-8c9f-9897cdcf3697'])`, r.users[3].ID),
		fmt.Sprintf(`insert into discussions values (55555, 'Answered Questions', 'This question is answered', 'closed', 'qanda', '%s', array['question','answered'],null,null)`, r.users[4].ID),
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

func (r *DiscussionRepositoryTestSuite) cleanup() error {
	queries := []string{
		"TRUNCATE TABLE discussions CASCADE",
	}
	return r.client.ExecQueries(r.ctx, queries)
}

func (r *DiscussionRepositoryTestSuite) TestCreate() {
	r.Run("should create a new discussion with type open ended and default state", func() {
		disc := &discussion.Discussion{
			Title:     "A New Discussion",
			Body:      "This is body",
			Type:      discussion.TypeOpenEnded,
			Labels:    []string{"label1", "label2"},
			Assets:    assetsToAssetIDs(r.assets),
			Assignees: usersToUserIDs(r.users)[:2],
			Owner:     r.users[len(r.users)-1],
		}
		id, err := r.repository.Create(r.ctx, disc)
		r.NoError(err)
		r.NotEmpty(id)
	})

	r.Run("should create a new discussion with empty asignee and assets", func() {
		disc := &discussion.Discussion{
			Title:  "A New Discussion",
			Body:   "This is body",
			Type:   discussion.TypeOpenEnded,
			Labels: []string{"label1", "label2"},
			Owner:  r.users[len(r.users)-1],
		}
		id, err := r.repository.Create(r.ctx, disc)
		r.NoError(err)
		r.NotEmpty(id)
	})

	r.Run("should return error when creating a new discussion with empty owner", func() {
		disc := &discussion.Discussion{
			Title:  "A New Discussion",
			Body:   "This is body",
			Type:   discussion.TypeOpenEnded,
			Labels: []string{"label1", "label2"},
		}
		id, err := r.repository.Create(r.ctx, disc)
		r.Error(err)
		r.Empty(id)
	})
}

func (r *DiscussionRepositoryTestSuite) TestGetAll() {
	r.Run("should return list of discussions if filter is valid", func() {
		testCases := []struct {
			description    string
			filter         discussion.Filter
			resultLength   int
			validateResult func(r *DiscussionRepositoryTestSuite, results []discussion.Discussion)
		}{
			{
				description:  "should successfully fetch all discussions",
				filter:       discussion.Filter{},
				resultLength: 5,
			},
			{
				description:  "should successfully fetch all discussions with all filters",
				filter:       discussion.Filter{Type: "all", State: "all"},
				resultLength: 5,
			},
			{
				description:  "should successfully fetch all discussions with assignee user 0 or owner user 0",
				filter:       discussion.Filter{Assignees: []string{r.users[0].ID}, Owner: r.users[0].ID, DisjointAssigneeOwner: true},
				resultLength: 2,
			},
			{
				description: "should limit with size",
				filter: discussion.Filter{
					Size: 1,
				},
				resultLength: 1,
			},
			{
				description: "should get all q&a type",
				filter: discussion.Filter{
					Type: discussion.TypeQAndA.String(),
				},
				resultLength: 2,
			},
			{
				description: "should only get closed q&a type",
				filter: discussion.Filter{
					Type:  discussion.TypeQAndA.String(),
					State: discussion.StateClosed.String(),
				},
				resultLength: 1,
				validateResult: func(r *DiscussionRepositoryTestSuite, results []discussion.Discussion) {
					r.Equal(results[0].Title, "Answered Questions")
				},
			},
			{
				description: "should get all discussions with label `question` and `data`",
				filter: discussion.Filter{
					Labels: []string{"question", "data"},
				},
				resultLength: 1,
				validateResult: func(r *DiscussionRepositoryTestSuite, results []discussion.Discussion) {
					r.Equal(results[0].ID, "22222")
				},
			},
			{
				description: "should get all discussions with a specific assignee",
				filter: discussion.Filter{
					Assignees: []string{"ef1f0896-785a-47a0-8c9f-9897cdcf3697"},
				},
				resultLength: 2,
			},
			{
				description: "should get all discussions ascendingly sorted by created_at",
				filter: discussion.Filter{
					SortBy:        "created_at",
					SortDirection: "asc",
				},
				resultLength: 5,
				validateResult: func(r *DiscussionRepositoryTestSuite, results []discussion.Discussion) {
					r.Equal(results[0].ID, "11111")
					r.Equal(results[1].ID, "22222")
					r.Equal(results[2].ID, "33333")
					r.Equal(results[3].ID, "44444")
					r.Equal(results[4].ID, "55555")
				},
			},
			{
				description: "should get all discussions with a specific assetid",
				filter: discussion.Filter{
					Assets: []string{"6663c3c7-62cf-41db-b0e1-4d443525f6d4"},
				},
				resultLength: 1,
				validateResult: func(r *DiscussionRepositoryTestSuite, results []discussion.Discussion) {
					r.Equal(results[0].ID, "22222")
				},
			},
			{
				description: "should get all discussions with a specific owner",
				filter: discussion.Filter{
					Owner: r.users[4].ID,
				},
				resultLength: 1,
				validateResult: func(r *DiscussionRepositoryTestSuite, results []discussion.Discussion) {
					r.Equal(results[0].Owner.ID, r.users[4].ID)
				},
			},
			{
				description: "should get all discussions with all filter populated",
				filter: discussion.Filter{
					Type:      discussion.TypeQAndA.String(),
					State:     discussion.StateOpen.String(),
					Labels:    []string{"wondering"},
					Assignees: []string{"59bf4219-433e-4e38-ad72-ebab70c7ee7a"},
					Assets:    []string{"6663c3c7-62cf-41db-b0e1-4d443525f6d4"},
					Owner:     r.users[1].ID,
				},
				resultLength: 1,
				validateResult: func(r *DiscussionRepositoryTestSuite, results []discussion.Discussion) {
					r.Equal(results[0].Owner.ID, r.users[1].ID)
				},
			},
		}

		for _, testCase := range testCases {
			r.Run(testCase.description, func() {
				dscs, err := r.repository.GetAll(r.ctx, testCase.filter)
				r.NoError(err)
				r.Len(dscs, testCase.resultLength)

				if testCase.validateResult != nil {
					testCase.validateResult(r, dscs)
				}
			})
		}
	})
}

func (r *DiscussionRepositoryTestSuite) TestGet() {
	testCases := []struct {
		description    string
		id             string
		validateResult func(r *DiscussionRepositoryTestSuite, result discussion.Discussion, err error)
	}{
		{
			description: "should successfully fetch all discussions",
			id:          "777",
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion, err error) {
				r.ErrorAs(err, new(discussion.NotFoundError))
				r.Empty(result)
			},
		},
		{
			description: "should return error if id is not serial",
			id:          "some-id",
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion, err error) {
				r.Error(err)
				r.Empty(result)
			},
		},
		{
			description: "should return discussion with a specific id",
			id:          "11111",
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion, err error) {
				r.NoError(err)
				r.Equal(result.ID, "11111")
			},
		},
		{
			description: "should return not found if id not found",
			id:          "99999",
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion, err error) {
				r.Error(discussion.NotFoundError{DiscussionID: "99999"})
				r.Empty(result)
			},
		},
	}

	for _, testCase := range testCases {
		r.Run(testCase.description, func() {
			dsc, err := r.repository.Get(r.ctx, testCase.id)
			testCase.validateResult(r, dsc, err)
		})
	}

	r.Run("discussion with deleted owner information would return empty owner", func() {
		queries := []string{
			`insert into discussions values (123, 'discussion with deleted user', 'discussion with deleted user', 1, 2, '59bf4219-433e-4e38-ad72-ebab70c7ee7a', array['kafka','topic','question'],null,array['59bf4219-433e-4e38-ad72-ebab70c7ee7a'])`,
		}
		err := r.client.ExecQueries(r.ctx, queries)
		r.NoError(err)
		dsc, err := r.repository.Get(r.ctx, "123")
		r.NoError(err)
		r.Empty(dsc.Owner)
	})
}

func (r *DiscussionRepositoryTestSuite) TestPatch() {
	testCases := []struct {
		description    string
		disc           *discussion.Discussion
		err            error
		validateResult func(r *DiscussionRepositoryTestSuite, result discussion.Discussion)
	}{
		{
			description: "should successfully patch title and body",
			disc:        &discussion.Discussion{ID: "11111", Title: "patched title", Body: "patched body"},
			err:         nil,
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion) {
				r.Equal("patched title", result.Title)
				r.Equal("patched body", result.Body)
			},
		},
		{
			description: "should successfully patch assignees",
			disc:        &discussion.Discussion{ID: "11111", Assignees: []string{"a", "b", "c"}},
			err:         nil,
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion) {
				r.Equal([]string{"a", "b", "c"}, result.Assignees)
			},
		},
		{
			description: "should successfully patch assets",
			disc:        &discussion.Discussion{ID: "11111", Assets: []string{"d", "e", "f"}},
			err:         nil,
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion) {
				r.Equal([]string{"d", "e", "f"}, result.Assets)
			},
		},
		{
			description: "should successfully patch assets empty",
			disc:        &discussion.Discussion{ID: "11111", Assets: []string{}},
			err:         nil,
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion) {
				r.Equal([]string(nil), result.Assets)
			},
		},
		{
			description: "should successfully patch labels empty",
			disc:        &discussion.Discussion{ID: "11111", Labels: []string{}},
			err:         nil,
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion) {
				r.Equal([]string(nil), result.Labels)
			},
		},
		{
			description: "should successfully patch assignees empty",
			disc:        &discussion.Discussion{ID: "11111", Assignees: []string{}},
			err:         nil,
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion) {
				r.Equal([]string(nil), result.Assignees)
			},
		}, {
			description: "should return error if discussion id does not exist",
			disc:        &discussion.Discussion{ID: "999", Assignees: []string{}},
			err:         discussion.NotFoundError{DiscussionID: "999"},
			validateResult: func(r *DiscussionRepositoryTestSuite, result discussion.Discussion) {
				r.Empty(result)
			},
		},
	}

	for _, testCase := range testCases {
		r.Run(testCase.description, func() {
			err := r.repository.Patch(r.ctx, testCase.disc)
			r.Equal(testCase.err, err)

			dsc, err := r.repository.Get(r.ctx, testCase.disc.ID)
			r.Equal(err, testCase.err)
			testCase.validateResult(r, dsc)
		})
	}
}

func TestDiscussionRepository(t *testing.T) {
	suite.Run(t, &DiscussionRepositoryTestSuite{})
}
