package handlersv1beta1

import (
	"context"
	"fmt"
	"github.com/odpf/compass/core/namespace"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestSearch(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
		ns       = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "base-test",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.SearchAssetsRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService, *mocks.NamespaceService)
		PostCheck    func(resp *compassv1beta1.SearchAssetsResponse) error
	}

	var testCases = []testCase{
		{
			Description:  "should return invalid argument if 'text' parameter is empty or missing",
			ExpectStatus: codes.InvalidArgument,
			Request:      &compassv1beta1.SearchAssetsRequest{},
		},
		{
			Description: "should report internal server if asset searcher fails",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text:         "test",
				NamespaceUrn: ns.ID.String(),
			},
			Setup: func(ctx context.Context, as *mocks.AssetService, nss *mocks.NamespaceService) {
				err := fmt.Errorf("service unavailable")
				as.EXPECT().SearchAssets(ctx, mock.AnythingOfType("asset.SearchConfig")).
					Return([]asset.SearchResult{}, err)
				nss.EXPECT().GetByID(ctx, ns.ID).Return(ns, nil)
			},
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should pass filter to search config format",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text: "resource",
				Filter: map[string]string{
					"data.landscape": "th",
					"type":           "topic",
					"service":        "kafka,rabbitmq",
				},
				NamespaceUrn: ns.ID.String(),
			},
			Setup: func(ctx context.Context, as *mocks.AssetService, nss *mocks.NamespaceService) {
				cfg := asset.SearchConfig{
					Text: "resource",
					Filters: map[string][]string{
						"type":           {"topic"},
						"service":        {"kafka", "rabbitmq"},
						"data.landscape": {"th"},
					},
					Namespace: ns,
				}

				as.EXPECT().SearchAssets(ctx, cfg).Return([]asset.SearchResult{}, nil)
				nss.EXPECT().GetByID(ctx, ns.ID).Return(ns, nil)
			},
		},
		{
			Description: "should pass queries to search config format",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text: "resource",
				Filter: map[string]string{
					"data.landscape": "th",
					"type":           "topic",
					"service":        "kafka,rabbitmq",
				},
				Query: map[string]string{
					"data.columns.name": "timestamp",
					"owners.email":      "john.doe@email.com",
				},
				NamespaceUrn: ns.ID.String(),
			},
			Setup: func(ctx context.Context, as *mocks.AssetService, nss *mocks.NamespaceService) {
				cfg := asset.SearchConfig{
					Text: "resource",
					Filters: map[string][]string{
						"type":           {"topic"},
						"service":        {"kafka", "rabbitmq"},
						"data.landscape": {"th"},
					},
					Queries: map[string]string{
						"data.columns.name": "timestamp",
						"owners.email":      "john.doe@email.com",
					},
					Namespace: ns,
				}

				as.EXPECT().SearchAssets(ctx, cfg).Return([]asset.SearchResult{}, nil)
				nss.EXPECT().GetByID(ctx, ns.ID).Return(ns, nil)
			},
		},
		{
			Description: "should return the matched documents",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text:         "test",
				NamespaceUrn: ns.ID.String(),
			},
			Setup: func(ctx context.Context, as *mocks.AssetService, nss *mocks.NamespaceService) {
				cfg := asset.SearchConfig{
					Text:      "test",
					Filters:   make(map[string][]string),
					Queries:   map[string]string(nil),
					Namespace: ns,
				}
				response := []asset.SearchResult{
					{
						Type:        "test",
						ID:          "test-resource",
						Description: "some description",
						Service:     "test-service",
						Labels: map[string]string{
							"entity":    "odpf",
							"landscape": "id",
						},
					},
				}
				as.EXPECT().SearchAssets(ctx, cfg).Return(response, nil)
				nss.EXPECT().GetByID(ctx, ns.ID).Return(ns, nil)
			},
			PostCheck: func(resp *compassv1beta1.SearchAssetsResponse) error {
				expected := &compassv1beta1.SearchAssetsResponse{
					Data: []*compassv1beta1.Asset{
						{
							Id:          "test-resource",
							Description: "some description",
							Service:     "test-service",
							Type:        "test",
							Labels: map[string]string{
								"entity":    "odpf",
								"landscape": "id",
							},
						},
					},
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
		{
			Description: "should return the requested number of assets",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text:         "resource",
				Size:         10,
				NamespaceUrn: ns.ID.String(),
			},
			Setup: func(ctx context.Context, as *mocks.AssetService, nss *mocks.NamespaceService) {
				cfg := asset.SearchConfig{
					Text:       "resource",
					MaxResults: 10,
					Filters:    make(map[string][]string),
					Queries:    map[string]string(nil),
					Namespace:  ns,
				}

				var results []asset.SearchResult
				for i := 0; i < cfg.MaxResults; i++ {
					urn := fmt.Sprintf("resource-%d", i+1)
					name := fmt.Sprintf("resource %d", i+1)
					r := asset.SearchResult{
						ID:          urn,
						Type:        "table",
						Description: name,
						Service:     "kafka",
						Labels: map[string]string{
							"landscape": "id",
							"entity":    "odpf",
						},
					}

					results = append(results, r)
				}

				as.EXPECT().SearchAssets(ctx, cfg).Return(results, nil)
				nss.EXPECT().GetByID(ctx, ns.ID).Return(ns, nil)
			},
			PostCheck: func(resp *compassv1beta1.SearchAssetsResponse) error {
				expectedSize := 10
				actualSize := len(resp.Data)
				if expectedSize != actualSize {
					return fmt.Errorf("expected search request to return %d results, returned %d results instead", expectedSize, actualSize)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.AssetService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc, mockNamespaceSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)
			defer mockNamespaceSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, mockSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.SearchAssets(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestSuggest(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
		ns       = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "base-test",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.SuggestAssetsRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService, *mocks.NamespaceService)
		PostCheck    func(resp *compassv1beta1.SuggestAssetsResponse) error
	}

	var testCases = []testCase{
		{
			Description:  "should return invalid arguments if 'text' parameter is empty or missing",
			ExpectStatus: codes.InvalidArgument,
			Request:      &compassv1beta1.SuggestAssetsRequest{},
		},
		{
			Description: "should report internal server error if searcher fails",
			Request: &compassv1beta1.SuggestAssetsRequest{
				Text:         "test",
				NamespaceUrn: ns.ID.String(),
			},
			Setup: func(ctx context.Context, as *mocks.AssetService, nss *mocks.NamespaceService) {
				cfg := asset.SearchConfig{
					Text:      "test",
					Namespace: ns,
				}
				as.EXPECT().SuggestAssets(ctx, cfg).Return([]string{}, fmt.Errorf("service unavailable"))
				nss.EXPECT().GetByID(ctx, ns.ID).Return(ns, nil)
			},
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return suggestions",
			Request: &compassv1beta1.SuggestAssetsRequest{
				Text:         "test",
				NamespaceUrn: ns.ID.String(),
			},
			Setup: func(ctx context.Context, as *mocks.AssetService, nss *mocks.NamespaceService) {
				cfg := asset.SearchConfig{
					Text:      "test",
					Namespace: ns,
				}
				response := []string{
					"test",
					"test2",
					"t est",
					"t_est",
				}

				as.EXPECT().SuggestAssets(ctx, cfg).Return(response, nil)
				nss.EXPECT().GetByID(ctx, ns.ID).Return(ns, nil)
			},
			PostCheck: func(resp *compassv1beta1.SuggestAssetsResponse) error {
				expected := &compassv1beta1.SuggestAssetsResponse{
					Data: []string{
						"test",
						"test2",
						"t est",
						"t_est",
					},
				}
				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.AssetService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc, mockNamespaceSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)
			defer mockNamespaceSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, mockSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.SuggestAssets(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}
