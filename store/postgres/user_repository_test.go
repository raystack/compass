package postgres_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.UserRepository
}

func (r *UserRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewNoop()
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.repository, err = postgres.NewUserRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *UserRepositoryTestSuite) TearDownSuite() {
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

func (r *UserRepositoryTestSuite) TestCreate() {
	r.Run("return no error if succesfully create user", func() {
		user := getUser("user@odpf.io")
		id, err := r.repository.Create(r.ctx, user)
		r.NoError(err)
		r.Equal(lengthOfString(id), 36) // uuid
	})

	r.Run("return ErrNoUserInformation if user is nil", func() {
		id, err := r.repository.Create(r.ctx, nil)
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if user is already exist", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		ud := getUser("user@odpf.io")
		id, err := r.repository.Create(r.ctx, ud)
		r.NoError(err)
		r.Equal(lengthOfString(id), 36) // uuid

		id, err = r.repository.Create(r.ctx, ud)
		r.ErrorAs(err, new(user.DuplicateRecordError))
		r.Empty(id)
	})

	r.Run("return invalid error if field in user is empty", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		ud := &user.User{}
		id, err := r.repository.Create(r.ctx, ud)
		r.ErrorIs(err, user.InvalidError{})
		r.Empty(id)
	})
}

func (r *UserRepositoryTestSuite) TestCreateWithTx() {
	validUser := &user.User{
		Email:    "userWithTx@odpf.io",
		Provider: "columbus",
	}
	r.Run("return no error if succesfully create user", func() {
		var id string
		err := r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, validUser)
			return err
		})
		r.Equal(lengthOfString(id), 36) // uuid
		r.NoError(err)
	})

	r.Run("return ErrNilUser if user is nil", func() {
		var id string
		err := r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, nil)
			return err
		})
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if user is already exist", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		id, err := r.repository.Create(r.ctx, validUser)
		r.NoError(err)
		r.Equal(lengthOfString(id), 36) // uuid

		err = r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, validUser)
			return err
		})
		r.ErrorAs(err, new(user.DuplicateRecordError))
		r.Empty(id)
	})
}

func (r *UserRepositoryTestSuite) TestGetID() {
	r.Run("return empty string and ErrNotFound if email not found in DB", func() {
		uid, err := r.repository.GetID(r.ctx, "random")
		r.ErrorAs(err, new(user.NotFoundError))
		r.Empty(uid)
	})

	r.Run("return non empty id if email found in DB", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		user := getUser("user@odpf.io")
		id, err := r.repository.Create(r.ctx, user)
		r.NoError(err)
		r.Equal(lengthOfString(id), 36) // uuid

		uid, err := r.repository.GetID(r.ctx, user.Email)
		r.NoError(err)
		r.NotEmpty(uid)
	})
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, &UserRepositoryTestSuite{})
}
