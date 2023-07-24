package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/compass/internal/testutils"
	"github.com/goto/salt/log"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	logLevelDebug       = "debug"
	defaultProviderName = "shield"
	defaultGetMaxSize   = 7
)

func newTestClient(t *testing.T, logger log.Logger) (*postgres.Client, error) {
	t.Helper()

	port, err := testutils.RunTestPG(t, logger)
	if err != nil {
		return nil, err
	}

	pgClient, err := postgres.NewClient(postgres.Config{
		Host:     testutils.PGHost,
		Port:     port,
		Name:     testutils.PGName,
		User:     testutils.PGUsername,
		Password: testutils.PGPassword,
	})
	if err != nil {
		return nil, err
	}

	if err := testutils.RunMigrationsWithClient(t, pgClient); err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		if err := pgClient.Close(); err != nil {
			t.Fatal(err)
		}
	})

	return pgClient, nil
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
