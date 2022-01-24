package postgres_test

import (
	"context"
	"fmt"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/odpf/columbus/store/postgres"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
)

var (
	pgConfig = postgres.Config{
		Host:     "localhost",
		User:     "test_user",
		Password: "test_pass",
		Name:     "test_db",
	}
)

func newTestClient(logger *logrus.Logger) (*postgres.Client, *dockertest.Pool, *dockertest.Resource, error) {

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
		return nil, nil, nil, fmt.Errorf("Could not create dockertest pool: %w", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Could not start resource: %w", err)
	}

	pgConfig.Port, err = strconv.Atoi(resource.GetPort("5432/tcp"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse external port of container to int: %w", err)
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

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	var pgClient *postgres.Client
	if err = pool.Retry(func() error {
		pgClient, err = postgres.NewClient(pgConfig)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, nil, fmt.Errorf("Could not connect to docker: %w", err)
	}

	err = setup(context.Background(), pgClient)
	if err != nil {
		logger.Fatal(err)
	}
	return pgClient, pool, resource, nil
}

func purgeDocker(pool *dockertest.Pool, resource *dockertest.Resource) error {
	if err := pool.Purge(resource); err != nil {
		return fmt.Errorf("Could not purge resource: %w", err)
	}
	return nil
}

func setup(ctx context.Context, client *postgres.Client) (err error) {
	var queries = []string{
		"DROP SCHEMA public CASCADE",
		"CREATE SCHEMA public",
	}
	err = client.ExecQueries(ctx, queries)
	if err != nil {
		return
	}

	err = client.Migrate(pgConfig)
	return
}
