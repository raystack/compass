package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/compass/internal/testutils"
	"github.com/goto/salt/log"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	repository *postgres.UserRepository
}

func (r *UserRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewNoop()
	r.client, err = newTestClient(r.T(), logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.repository, err = postgres.NewUserRepository(r.client)
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
		user := getUser("user@gotocompany.com")
		id, err := r.repository.Create(r.ctx, user)
		r.NotEmpty(id)
		r.NoError(err)
	})

	r.Run("return ErrNoUserInformation if user is nil", func() {
		id, err := r.repository.Create(r.ctx, nil)
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if user is already exist", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		ud := getUser("user@gotocompany.com")
		id, err := r.repository.Create(r.ctx, ud)
		r.NoError(err)
		r.NotEmpty(id)

		id, err = r.repository.Create(r.ctx, ud)
		r.ErrorAs(err, new(user.DuplicateRecordError))
		r.Empty(id)
	})
}

func (r *UserRepositoryTestSuite) TestCreateWithTx() {
	validUserWithoutUUID := &user.User{
		Email:    "userWithTx@gotocompany.com",
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
			id, err = r.repository.CreateWithTx(r.ctx, tx, validUserWithoutUUID)
			return err
		})
		r.NotEmpty(id)
		r.NoError(err)
	})

	r.Run("return no error if succesfully create user without email", func() {
		var id string
		err := r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, validUserWithoutEmail)
			return err
		})
		r.NotEmpty(id)
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
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		id, err := r.repository.Create(r.ctx, validUserWithoutUUID)
		r.NoError(err)
		r.NotEmpty(id)

		err = r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, validUserWithoutUUID)
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
			err := testutils.RunMigrationsWithClient(r.T(), r.client)
			r.NoError(err)

			user := getUser("use-getbyemail@gotocompany.com")
			id, err := r.repository.Create(r.ctx, user)
			r.NoError(err)
			r.NotEmpty(id)

			usr, err := r.repository.GetByEmail(r.ctx, user.Email)
			r.NoError(err)
			r.NotEmpty(usr)
		})
	})

	r.Run("by email with tx, return the user created in the tx", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		u := getUser("use-getbyemail@gotocompany.com")
		err = r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			id, err := r.repository.CreateWithTx(r.ctx, tx, u)
			r.NoError(err)

			_, err = r.repository.GetByEmail(r.ctx, u.Email)
			r.ErrorAs(err, new(user.NotFoundError))

			u, err := r.repository.GetByEmailWithTx(r.ctx, tx, u.Email)
			r.NoError(err)
			r.Equal(id, u.ID)

			return nil
		})
		r.NoError(err)
	})

	r.Run("by uuid", func() {
		r.Run("return empty string and ErrNotFound if uuid not found in DB", func() {
			usr, err := r.repository.GetByUUID(r.ctx, "random")
			r.ErrorAs(err, new(user.NotFoundError))
			r.Empty(usr)
		})

		r.Run("return non empty user if email found in DB", func() {
			err := testutils.RunMigrationsWithClient(r.T(), r.client)
			r.NoError(err)

			user := getUser("use-getbyuuid@gotocompany.com")
			id, err := r.repository.Create(r.ctx, user)
			r.NoError(err)
			r.NotEmpty(id)

			usr, err := r.repository.GetByUUID(r.ctx, user.UUID)
			r.NoError(err)
			r.NotEmpty(usr)
		})
	})

	r.Run("by UUID with tx, return the user created in the tx", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		u := getUser("use-getbyuuid@gotocompany.com")
		err = r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			id, err := r.repository.CreateWithTx(r.ctx, tx, u)
			r.NoError(err)

			_, err = r.repository.GetByUUID(r.ctx, u.UUID)
			r.ErrorAs(err, new(user.NotFoundError))

			u, err := r.repository.GetByUUIDWithTx(r.ctx, tx, u.UUID)
			r.NoError(err)
			r.Equal(id, u.ID)

			return nil
		})
		r.NoError(err)
	})
}

func (r *UserRepositoryTestSuite) TestUpsertByEmail() {
	r.Run("return ErrNoUserInformation if user is nil", func() {
		id, err := r.repository.UpsertByEmail(r.ctx, nil)
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if record already exist", func() {
		usr := &user.User{UUID: uuid.NewString(), Email: "dummy@gotocompany.com"}

		err := r.insertEmail(usr.Email)
		r.NoError(err)

		usr.UUID = uuid.NewString()
		id, err := r.repository.UpsertByEmail(r.ctx, usr)
		r.NoError(err)
		r.NotEmpty(id)

		id, err = r.repository.UpsertByEmail(r.ctx, usr)
		r.ErrorIs(err, user.DuplicateRecordError{UUID: usr.UUID, Email: usr.Email})
		r.Empty(id)
	})

	r.Run("new row is inserted with uuid and email if user not exist", func() {
		usr := &user.User{UUID: uuid.NewString(), Email: "user-upsert-1@gotocompany.com"}
		id, err := r.repository.UpsertByEmail(r.ctx, usr)
		r.NoError(err)
		r.NotEmpty(id)

		gotUser, err := r.repository.GetByUUID(r.ctx, usr.UUID)
		r.NoError(err)
		r.Equal(gotUser.UUID, usr.UUID)
		r.Equal(gotUser.Email, usr.Email)
	})

	r.Run("new row is inserted with uuid only if user not exist", func() {
		usr := &user.User{UUID: uuid.NewString()}
		id, err := r.repository.UpsertByEmail(r.ctx, usr)
		r.NoError(err)
		r.NotEmpty(id)

		gotUser, err := r.repository.GetByUUID(r.ctx, usr.UUID)
		r.NoError(err)
		r.Equal(gotUser.UUID, usr.UUID)
		r.Equal(gotUser.Email, usr.Email)
	})

	r.Run("upserting existing row with empty uuid is upserted with uuid and email", func() {
		usr := &user.User{Email: "user-upsert-2@gotocompany.com"}

		err := r.insertEmail(usr.Email)
		r.NoError(err)

		usr.UUID = uuid.NewString()
		id, err := r.repository.UpsertByEmail(r.ctx, usr)
		r.NoError(err)
		r.NotEmpty(id)

		gotUser, err := r.repository.GetByUUID(r.ctx, usr.UUID)
		r.NoError(err)
		r.Equal(gotUser.UUID, usr.UUID)
		r.Equal(gotUser.Email, usr.Email)
	})

	r.Run("upserting existing row with non empty uuid would return error", func() {
		usr := &user.User{UUID: uuid.NewString(), Email: "user-upsert-3@gotocompany.com"}

		id, err := r.repository.Create(r.ctx, usr)
		r.NoError(err)
		r.NotEmpty(id)

		id, err = r.repository.UpsertByEmail(r.ctx, usr)
		r.Error(err)
		r.Empty(id)
	})
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, &UserRepositoryTestSuite{})
}
