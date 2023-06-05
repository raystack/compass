package postgres_test

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/salt/log"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	logLevelDebug       = "debug"
	defaultProviderName = "shield"
	defaultGetMaxSize   = 7
)

var pgConfig = postgres.Config{
	Host:     "localhost",
	User:     "test_user",
	Password: "test_pass",
	Name:     "test_db",
}

func newTestClient(logger log.Logger) (*postgres.Client, *dockertest.Pool, *dockertest.Resource, error) {
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
		return nil, nil, nil, fmt.Errorf("could not create dockertest pool: %w", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	pgConfig.Port, err = strconv.Atoi(resource.GetPort("5432/tcp"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse external port of container to int: %w", err)
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
			logger.Fatal("could not connect to postgres container log output", "error", err)
		}
		defer func() {
			err = logWaiter.Close()
			if err != nil {
				logger.Fatal("could not close container log", "error", err)
			}

			err = logWaiter.Wait()
			if err != nil {
				logger.Fatal("could not wait for container log to close", "error", err)
			}
		}()
	}

	// Tell docker to hard kill the container in 120 seconds
	if err := resource.Expire(120); err != nil {
		return nil, nil, nil, err
	}

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
		return nil, nil, nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	err = setup(context.Background(), pgClient)
	if err != nil {
		logger.Fatal("failed to setup and migrate DB", "error", err)
	}
	return pgClient, pool, resource, nil
}

func purgeDocker(pool *dockertest.Pool, resource *dockertest.Resource) error {
	if err := pool.Purge(resource); err != nil {
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}

func setup(ctx context.Context, client *postgres.Client) error {
	queries := []string{
		"DROP SCHEMA public CASCADE",
		"CREATE SCHEMA public",
	}

	if err := client.ExecQueries(ctx, queries); err != nil {
		return err
	}

	return client.Migrate(pgConfig)
}

// helper functions
func createUser(userRepo user.Repository, email string) (string, error) {
	user := getUser(email)
	id, err := userRepo.Create(context.Background(), user)
	if err != nil {
		return "", err
	}
	return id, nil
}

func createAsset(assetRepo asset.Repository, updaterID, ownerEmail, assetURN, assetType string) (*asset.Asset, error) {
	ast := getAsset(ownerEmail, assetURN, assetType)
	ast.UpdatedBy.ID = updaterID
	id, err := assetRepo.Upsert(context.Background(), ast)
	if err != nil {
		return nil, err
	}
	ast.ID = id
	return ast, nil
}

func getAsset(ownerEmail, assetURN, assetType string) *asset.Asset {
	return &asset.Asset{
		URN:     assetURN,
		Type:    asset.Type(assetType),
		Service: "bigquery",
		Owners: []user.User{
			{
				Email: ownerEmail,
			},
		},
		UpdatedBy: user.User{
			Email: ownerEmail,
		},
	}
}

func getUser(email string) *user.User {
	timestamp := time.Now().UTC()
	return &user.User{
		UUID:      uuid.NewString(),
		Email:     email,
		Provider:  defaultProviderName,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}
}

func createUsers(userRepo user.Repository, num int) ([]user.User, error) {
	var users []user.User
	for i := 0; i < num; i++ {
		email := fmt.Sprintf("user-test-%d@gotocompany.com", i+1)
		user1 := user.User{UUID: uuid.NewString(), Email: email, Provider: defaultProviderName}
		var err error
		user1.ID, err = userRepo.Create(context.Background(), &user1)
		if err != nil {
			return nil, err
		}
		users = append(users, user1)
	}
	return users, nil
}

func createAssets(assetRepo asset.Repository, users []user.User, astType asset.Type) ([]asset.Asset, error) {
	var aa []asset.Asset
	count := 0
	for _, usr := range users {
		var ast *asset.Asset
		count += 1
		assetURN := fmt.Sprintf("asset-urn-%d", count)
		ast, err := createAsset(assetRepo, usr.ID, usr.Email, assetURN, astType.String())
		if err != nil {
			return nil, err
		}
		aa = append(aa, *ast)
	}
	return aa, nil
}

func usersToUserIDs(users []user.User) []string {
	ids := make([]string, 0, len(users))
	for _, us := range users {
		ids = append(ids, us.ID)
	}
	return ids
}

func assetsToAssetIDs(assets []asset.Asset) []string {
	ids := make([]string, 0, len(assets))
	for _, as := range assets {
		ids = append(ids, as.ID)
	}
	return ids
}
