package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	logLevelDebug = "debug"
	PGHost        = "localhost"
	PGUsername    = "test_user"
	PGPassword    = "test_pass"
	PGName        = "test_db"
)

func RunTestPG(t *testing.T, logger log.Logger) (int, error) {
	t.Helper()

	opts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		Env: []string{
			"POSTGRES_PASSWORD=" + PGPassword,
			"POSTGRES_USER=" + PGUsername,
			"POSTGRES_DB=" + PGName,
		},
	}

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		return 0, fmt.Errorf("new test PG: create dockertest pool: %w", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return 0, fmt.Errorf("new test PG: start resource: %w", err)
	}

	port, err := strconv.Atoi(resource.GetPort("5432/tcp"))
	if err != nil {
		return 0, fmt.Errorf("new test PG: parse external port of container to int: %w", err)
	}

	// attach terminal logger to container if exists
	// for debugging purpose
	if logger.Level() == logLevelDebug {
		logWaiter, err := pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
			Container:    resource.Container.ID,
			OutputStream: logger.Writer(),
			ErrorStream:  logger.Writer(),
			Stderr:       true,
			Stdout:       true,
			Stream:       true,
		})
		if err != nil {
			return 0, fmt.Errorf("new test PG: connect to postgres container log output: %w", err)
		}
		defer func() {
			if err := logWaiter.Close(); err != nil {
				logger.Error("could not close container log", "error", err)
			}

			if err := logWaiter.Wait(); err != nil {
				logger.Error("could not wait for container log to close", "error", err)
			}
		}()
	}

	// Tell docker to hard kill the container in 120 seconds
	if err := resource.Expire(120); err != nil {
		return 0, err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	if err := pool.Retry(func() error {
		db, err := sql.Open("pgx", fmt.Sprintf(
			"dbname=%s user=%s password='%s' host=%s port=%d sslmode=disable",
			PGName, PGUsername, PGPassword, PGHost, port,
		))
		if err != nil {
			return err
		}

		defer db.Close()

		return db.Ping()
	}); err != nil {
		return 0, fmt.Errorf("could not connect to docker: %w", err)
	}

	t.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			t.Fatal(err)
		}
	})

	return port, nil
}

func RunMigrations(t *testing.T, db *sql.DB) error {
	t.Helper()

	return RunMigrationsWithClient(t, postgres.NewClientWithDB(db))
}

func RunMigrationsWithClient(t *testing.T, pgClient *postgres.Client) error {
	t.Helper()

	queries := []string{
		"DROP SCHEMA public CASCADE",
		"CREATE SCHEMA public",
	}
	if err := pgClient.ExecQueries(context.Background(), queries); err != nil {
		return err
	}

	return pgClient.Migrate()
}
