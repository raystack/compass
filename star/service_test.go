package star_test

import (
	"context"
	"testing"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStar(t *testing.T) {
	type testCase struct {
		Description  string
		ExpectResult string
		ExpectError  error
		UserID       string
		Starring     *star.Star
		Setup        func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository)
		PostCheck    func(t *testing.T, tc *testCase, id string) error
	}

	userID := "1234-5678"
	assetURN := "asset:urn"
	assetType := "table"
	assetID := "8765-4321"

	var testCases = []testCase{
		{
			Description:  "should return star invalid error if asset in star is invalid",
			ExpectResult: "",
			ExpectError:  star.InvalidError{},
			Setup:        func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {},
		},
		{
			Description:  "should return asset not found if GetIDByURN return error",
			ExpectResult: "",
			ExpectError:  asset.NotFoundError{},
			UserID:       userID,
			Starring:     &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return("", asset.NotFoundError{})
			},
		},
		{
			Description:  "should return star error if create star return error",
			ExpectResult: "",
			ExpectError:  star.NotFoundError{},
			UserID:       userID,
			Starring:     &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return(assetID, nil)
				sr.On("Create", mock.AnythingOfType("*context.emptyCtx"), userID, assetID).Return("", star.NotFoundError{})
			},
		},
		{
			Description:  "should return star id if succesfully stars an asset",
			ExpectResult: "star-id",
			ExpectError:  nil,
			UserID:       userID,
			Starring:     &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return(assetID, nil)
				sr.On("Create", mock.AnythingOfType("*context.emptyCtx"), userID, assetID).Return("star-id", nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			ar := new(mocks.AssetRepository)
			defer sr.AssertExpectations(t)
			defer ar.AssertExpectations(t)
			tc.Setup(&tc, sr, ar)

			svc := star.NewService(sr, ar)
			result, err := svc.Star(context.Background(), tc.UserID, tc.Starring)
			assert.Equal(t, result, tc.ExpectResult)
			assert.ErrorIs(t, err, tc.ExpectError)
		})
	}
}

func TestGetStargazersByURN(t *testing.T) {
	type testCase struct {
		Description  string
		ExpectResult []user.User
		ExpectError  error
		Starring     *star.Star
		Setup        func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository)
	}

	starCfg := star.Config{}
	assetURN := "asset:urn"
	assetType := "table"
	assetID := "8765-4321"

	var testCases = []testCase{
		{
			Description:  "should return star invalid error if asset in star is invalid",
			ExpectResult: nil,
			ExpectError:  star.InvalidError{},
			Setup:        func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {},
		},
		{
			Description:  "should return asset not found if GetIDByURN return error",
			ExpectResult: nil,
			ExpectError:  asset.NotFoundError{},
			Starring:     &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return("", asset.NotFoundError{})
			},
		},
		{
			Description:  "should return star error if get stargazers return error",
			ExpectResult: nil,
			ExpectError:  star.NotFoundError{},
			Starring:     &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return(assetID, nil)
				sr.On("GetStargazers", mock.AnythingOfType("*context.emptyCtx"), starCfg, assetID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return list of users if succesfully getting stargazers",
			ExpectResult: []user.User{{ID: "123"}},
			ExpectError:  nil,
			Starring:     &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return(assetID, nil)
				sr.On("GetStargazers", mock.AnythingOfType("*context.emptyCtx"), starCfg, assetID).Return([]user.User{{ID: "123"}}, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			ar := new(mocks.AssetRepository)
			defer sr.AssertExpectations(t)
			defer ar.AssertExpectations(t)
			tc.Setup(&tc, sr, ar)

			svc := star.NewService(sr, ar)
			result, err := svc.GetStargazersByURN(context.Background(), starCfg, tc.Starring)
			assert.Equal(t, result, tc.ExpectResult)
			assert.ErrorIs(t, err, tc.ExpectError)
		})
	}
}

func TestGetStargazersByID(t *testing.T) {
	type testCase struct {
		Description  string
		ExpectResult []user.User
		ExpectError  error
		AssetID      string
		Setup        func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository)
	}

	starCfg := star.Config{}
	assetID := "8765-4321"

	var testCases = []testCase{
		{
			Description:  "should return star invalid error if asset id is empty",
			ExpectResult: nil,
			ExpectError:  star.InvalidError{},
			Setup:        func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {},
		},
		{
			Description:  "should return star error if get stargazers return error",
			ExpectResult: nil,
			ExpectError:  star.NotFoundError{},
			AssetID:      assetID,
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				sr.On("GetStargazers", mock.AnythingOfType("*context.emptyCtx"), starCfg, assetID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return list of users if succesfully getting stargazers",
			ExpectResult: []user.User{{ID: "123"}},
			ExpectError:  nil,
			AssetID:      assetID,
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				sr.On("GetStargazers", mock.AnythingOfType("*context.emptyCtx"), starCfg, assetID).Return([]user.User{{ID: "123"}}, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			ar := new(mocks.AssetRepository)
			defer sr.AssertExpectations(t)
			defer ar.AssertExpectations(t)
			tc.Setup(&tc, sr, ar)

			svc := star.NewService(sr, ar)
			result, err := svc.GetStargazersByID(context.Background(), starCfg, tc.AssetID)
			assert.Equal(t, result, tc.ExpectResult)
			assert.ErrorIs(t, err, tc.ExpectError)
		})
	}
}

func TestGetAssetByUserID(t *testing.T) {
	type testCase struct {
		Description  string
		ExpectResult *asset.Asset
		ExpectError  error
		UserID       string
		Starring     *star.Star
		Setup        func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository)
		PostCheck    func(t *testing.T, tc *testCase, ast *asset.Asset) error
	}

	userID := "1234-5678"
	assetURN := "asset:urn"
	assetType := "table"
	assetID := "8765-4321"

	var testCases = []testCase{
		{
			Description:  "should return star invalid error if asset in star is invalid",
			ExpectResult: nil,
			ExpectError:  star.InvalidError{},
			Setup:        func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {},
		},
		{
			Description:  "should return asset not found if GetIDByURN return error",
			ExpectResult: nil,
			ExpectError:  asset.NotFoundError{},
			UserID:       userID,
			Starring:     &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return("", asset.NotFoundError{})
			},
		},
		{
			Description:  "should return star error if create star return error",
			ExpectResult: nil,
			ExpectError:  star.NotFoundError{},
			UserID:       userID,
			Starring:     &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return(assetID, nil)
				sr.On("GetAssetByUserID", mock.AnythingOfType("*context.emptyCtx"), userID, assetID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return star id if succesfully stars an asset",
			ExpectResult: &asset.Asset{ID: "123"},
			ExpectError:  nil,
			UserID:       userID,
			Starring:     &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return(assetID, nil)
				sr.On("GetAssetByUserID", mock.AnythingOfType("*context.emptyCtx"), userID, assetID).Return(&asset.Asset{ID: "123"}, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			ar := new(mocks.AssetRepository)
			defer sr.AssertExpectations(t)
			defer ar.AssertExpectations(t)
			tc.Setup(&tc, sr, ar)

			svc := star.NewService(sr, ar)
			result, err := svc.GetAssetByUserID(context.Background(), tc.UserID, tc.Starring)
			assert.Equal(t, result, tc.ExpectResult)
			assert.ErrorIs(t, err, tc.ExpectError)
		})
	}
}

func TestUnstar(t *testing.T) {
	type testCase struct {
		Description string
		ExpectError error
		UserID      string
		Starring    *star.Star
		Setup       func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository)
	}

	userID := "1234-5678"
	assetURN := "asset:urn"
	assetType := "table"
	assetID := "8765-4321"

	var testCases = []testCase{
		{
			Description: "should return star invalid error if asset in star is invalid",
			ExpectError: star.InvalidError{},
			Setup:       func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {},
		},
		{
			Description: "should return asset not found if GetIDByURN return error",
			ExpectError: asset.NotFoundError{},
			UserID:      userID,
			Starring:    &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return("", asset.NotFoundError{})
			},
		},
		{
			Description: "should return star error if create star return error",
			ExpectError: star.NotFoundError{},
			UserID:      userID,
			Starring:    &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return(assetID, nil)
				sr.On("Delete", mock.AnythingOfType("*context.emptyCtx"), userID, assetID).Return(star.NotFoundError{})
			},
		},
		{
			Description: "should return star id if succesfully stars an asset",
			ExpectError: nil,
			UserID:      userID,
			Starring:    &star.Star{Asset: asset.Asset{URN: assetURN, Type: asset.Type(assetType)}},
			Setup: func(tc *testCase, sr *mocks.StarRepository, ar *mocks.AssetRepository) {
				ar.On("GetIDByURN", mock.AnythingOfType("*context.emptyCtx"), &asset.Asset{URN: assetURN, Type: asset.Type(assetType)}).Return(assetID, nil)
				sr.On("Delete", mock.AnythingOfType("*context.emptyCtx"), userID, assetID).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			ar := new(mocks.AssetRepository)
			defer sr.AssertExpectations(t)
			defer ar.AssertExpectations(t)
			tc.Setup(&tc, sr, ar)

			svc := star.NewService(sr, ar)
			err := svc.Unstar(context.Background(), tc.UserID, tc.Starring)
			assert.ErrorIs(t, err, tc.ExpectError)
		})
	}
}
