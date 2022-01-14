package postgres_test

import (
	"fmt"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/store/postgres"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultDB = "postgres"
)

var (
	testDBClient *postgres.Client
	pgConfig     = postgres.Config{
		Host:     "localhost",
		User:     "test_user",
		Password: "test_pass",
		Name:     "test_db",
	}
)

func newTestClient(logger *logrus.Logger) (*dockertest.Pool, *dockertest.Resource, error) {

	opts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		Env: []string{
			"POSTGRES_PASSWORD=" + pgConfig.Password,
			"POSTGRES_USER=" + pgConfig.User,
			"POSTGRES_DB=" + pgConfig.Name,
		},
	}

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not create dockertest pool")
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not start resource")
	}

	// pgConfig.Host = "localhost"
	pgConfig.Port, err = strconv.Atoi(resource.GetPort("5432/tcp"))
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot parse external port of container to int")
	}

	// attach terminal logger to container if exists
	// for debugging purpose
	if logger.Level == logrus.DebugLevel {
		logWaiter, err := pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
			Container:    resource.Container.ID,
			OutputStream: logger.Writer(),
			ErrorStream:  logger.Writer(),
			Stderr:       true,
			Stdout:       true,
			Stream:       true,
		})
		if err != nil {
			logger.WithError(err).Fatal("Could not connect to postgres container log output")
		}
		defer func() {
			err = logWaiter.Close()
			if err != nil {
				logger.WithError(err).Error("Could not close container log")
			}

			err = logWaiter.Wait()
			if err != nil {
				logger.WithError(err).Error("Could not wait for container log to close")
			}
		}()
	}

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	testDBClient, err = postgres.NewClient(logger, pgConfig)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not create pg client")
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	if err = pool.Retry(func() error {
		testDBClient.Conn, err = sqlx.Connect("pgx", pgConfig.ConnectionURL().String())
		return err
	}); err != nil {
		return nil, nil, errors.Wrap(err, "Could not connect to docker")
	}

	err = setup()
	if err != nil {
		return nil, nil, errors.Wrap(err, "error migrating DB")
	}

	return pool, resource, nil
}

func purgeClient(pool *dockertest.Pool, resource *dockertest.Resource) error {
	if err := pool.Purge(resource); err != nil {
		return errors.Wrap(err, "Could not purge resource")
	}
	return nil
}

func setup() (err error) {
	const testDB = "test_db"
	var queries = []string{
		fmt.Sprintf("DROP SCHEMA public CASCADE"),
		fmt.Sprintf("CREATE SCHEMA public"),
	}
	err = execute(testDBClient.Conn, queries)
	if err != nil {
		return
	}

	err = testDBClient.Migrate()
	return
}

func execute(db *sqlx.DB, queries []string) (err error) {
	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			return
		}
	}
	return
}
