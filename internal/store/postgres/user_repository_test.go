package postgres_test

import (
	"context"
	"fmt"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/pkg/grpc_interceptor"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/raystack/compass/core/user"
	"github.com/raystack/compass/internal/store/postgres"
	"github.com/raystack/salt/log"
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
	ns         *namespace.Namespace
}

func (r *UserRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewNoop()
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		r.T().Fatal(err)
	}
	r.ns = &namespace.Namespace{
		ID:       uuid.New(),
		Name:     "umbrella",
		State:    namespace.SharedState,
		Metadata: nil,
	}
	r.ctx = grpc_interceptor.BuildContextWithNamespace(context.Background(), r.ns)
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

func (r *UserRepositoryTestSuite) insertEmail(email string) error {
	query := fmt.Sprintf("insert into users (email) values ('%s')", email)
	if err := r.client.ExecQueries(context.Background(), []string{
		query,
	}); err != nil {
		return err
	}
	return nil
}

func (r *UserRepositoryTestSuite) TestCreate() {
	r.Run("return no error if succesfully create user", func() {
		user := getUser("user@raystack.io")
		id, err := r.repository.Create(r.ctx, r.ns, user)
		r.NotEmpty(id)
		r.NoError(err)
	})

	r.Run("return ErrNoUserInformation if user is nil", func() {
		id, err := r.repository.Create(r.ctx, r.ns, nil)
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if user is already exist", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		ud := getUser("user@raystack.io")
		id, err := r.repository.Create(r.ctx, r.ns, ud)
		r.NoError(err)
		r.NotEmpty(id)

		id, err = r.repository.Create(r.ctx, r.ns, ud)
		r.ErrorAs(err, new(user.DuplicateRecordError))
		r.Empty(id)
	})
}

func (r *UserRepositoryTestSuite) TestCreateWithTx() {
	validUserWithoutUUID := &user.User{
		Email:    "userWithTx@raystack.io",
		Provider: "compass",
	}
	validUserWithoutEmail := &user.User{
		UUID:     "a-uuid",
		Provider: "compass",
	}
	r.Run("return no error if succesfully create user without uuid", func() {
		var id string
		err := r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, r.ns, validUserWithoutUUID)
			return err
		})
		r.NotEmpty(id)
		r.NoError(err)
	})

	r.Run("return no error if succesfully create user without email", func() {
		var id string
		err := r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, r.ns, validUserWithoutEmail)
			return err
		})
		r.NotEmpty(id)
		r.NoError(err)
	})

	r.Run("return ErrNilUser if user is nil", func() {
		var id string
		err := r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, r.ns, nil)
			return err
		})
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if user is already exist", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		id, err := r.repository.Create(r.ctx, r.ns, validUserWithoutUUID)
		r.NoError(err)
		r.NotEmpty(id)

		err = r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, r.ns, validUserWithoutUUID)
			return err
		})
		r.ErrorAs(err, new(user.DuplicateRecordError))
		r.Empty(id)
	})
}

func (r *UserRepositoryTestSuite) TestGetBy() {
	r.Run("by email", func() {
		r.Run("return empty string and ErrNotFound if email not found in DB", func() {
			usr, err := r.repository.GetByEmail(r.ctx, "random")
			r.ErrorAs(err, new(user.NotFoundError))
			r.Empty(usr)
		})

		r.Run("return non empty user if email found in DB", func() {
			err := setup(r.ctx, r.client)
			r.NoError(err)

			user := getUser("use-getbyemail@raystack.io")
			id, err := r.repository.Create(r.ctx, r.ns, user)
			r.NoError(err)
			r.NotEmpty(id)

			usr, err := r.repository.GetByEmail(r.ctx, user.Email)
			r.NoError(err)
			r.NotEmpty(usr)
		})
	})

	r.Run("by uuid", func() {
		r.Run("return empty string and ErrNotFound if uuid not found in DB", func() {
			usr, err := r.repository.GetByUUID(r.ctx, "random")
			r.ErrorAs(err, new(user.NotFoundError))
			r.Empty(usr)
		})

		r.Run("return non empty user if email found in DB", func() {
			err := setup(r.ctx, r.client)
			r.NoError(err)

			user := getUser("use-getbyuuid@raystack.io")
			id, err := r.repository.Create(r.ctx, r.ns, user)
			r.NoError(err)
			r.NotEmpty(id)

			usr, err := r.repository.GetByUUID(r.ctx, user.UUID)
			r.NoError(err)
			r.NotEmpty(usr)
		})
	})

}

func (r *UserRepositoryTestSuite) TestUpsertByEmail() {
	r.Run("return ErrNoUserInformation if user is nil", func() {
		id, err := r.repository.UpsertByEmail(r.ctx, r.ns, nil)
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if record already exist", func() {
		usr := &user.User{UUID: uuid.NewString(), Email: "dummy@raystack.io"}

		err := r.insertEmail(usr.Email)
		r.NoError(err)

		usr.UUID = uuid.NewString()
		id, err := r.repository.UpsertByEmail(r.ctx, r.ns, usr)
		r.NoError(err)
		r.NotEmpty(id)

		id, err = r.repository.UpsertByEmail(r.ctx, r.ns, usr)
		r.ErrorIs(err, user.DuplicateRecordError{UUID: usr.UUID, Email: usr.Email})
		r.Empty(id)
	})

	r.Run("new row is inserted with uuid and email if user not exist", func() {
		usr := &user.User{UUID: uuid.NewString(), Email: "user-upsert-1@raystack.io"}
		id, err := r.repository.UpsertByEmail(r.ctx, r.ns, usr)
		r.NoError(err)
		r.NotEmpty(id)

		gotUser, err := r.repository.GetByUUID(r.ctx, usr.UUID)
		r.NoError(err)
		r.Equal(gotUser.UUID, usr.UUID)
		r.Equal(gotUser.Email, usr.Email)
	})

	r.Run("new row is inserted with uuid only if user not exist", func() {
		usr := &user.User{UUID: uuid.NewString()}
		id, err := r.repository.UpsertByEmail(r.ctx, r.ns, usr)
		r.NoError(err)
		r.NotEmpty(id)

		gotUser, err := r.repository.GetByUUID(r.ctx, usr.UUID)
		r.NoError(err)
		r.Equal(gotUser.UUID, usr.UUID)
		r.Equal(gotUser.Email, usr.Email)
	})

	r.Run("upserting existing row with empty uuid is upserted with uuid and email", func() {
		usr := &user.User{Email: "user-upsert-2@raystack.io"}

		err := r.insertEmail(usr.Email)
		r.NoError(err)

		usr.UUID = uuid.NewString()
		id, err := r.repository.UpsertByEmail(r.ctx, r.ns, usr)
		r.NoError(err)
		r.NotEmpty(id)

		gotUser, err := r.repository.GetByUUID(r.ctx, usr.UUID)
		r.NoError(err)
		r.Equal(gotUser.UUID, usr.UUID)
		r.Equal(gotUser.Email, usr.Email)
	})

	r.Run("upserting existing row with non empty uuid would return error", func() {
		usr := &user.User{UUID: uuid.NewString(), Email: "user-upsert-3@raystack.io"}

		id, err := r.repository.Create(r.ctx, r.ns, usr)
		r.NoError(err)
		r.NotEmpty(id)

		id, err = r.repository.UpsertByEmail(r.ctx, r.ns, usr)
		r.Error(err)
		r.Empty(id)
	})
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, &UserRepositoryTestSuite{})
}
