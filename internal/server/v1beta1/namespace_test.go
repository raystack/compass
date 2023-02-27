package handlersv1beta1

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/odpf/compass/core/namespace"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"testing"
)

func TestAPIServer_ListNamespaces(t *testing.T) {
	var (
		userID           = uuid.NewString()
		userUUID         = uuid.NewString()
		mockedNamespaces = []*namespace.Namespace{
			{
				ID:       uuid.New(),
				Name:     "tenant-1",
				State:    namespace.SharedState,
				Metadata: nil,
			},
			{
				ID:       uuid.New(),
				Name:     "tenant-2",
				State:    namespace.DedicatedState,
				Metadata: nil,
			},
		}
	)
	type testCase struct {
		name         string
		Request      *compassv1beta1.ListNamespacesRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.NamespaceService)
		PostCheck    func(resp *compassv1beta1.ListNamespacesResponse) error
	}
	var testCases = []testCase{
		{
			name:         "list namespace items successfully",
			Request:      &compassv1beta1.ListNamespacesRequest{},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().List(ctx).Return(mockedNamespaces, nil)
			},
			PostCheck: func(resp *compassv1beta1.ListNamespacesResponse) error {
				assert.Equal(t, len(resp.Namespaces), len(mockedNamespaces))
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockNamespaceSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockNamespaceSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, nil, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.ListNamespaces(ctx, tc.Request)
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

func TestAPIServer_GetNamespaces(t *testing.T) {
	var (
		userID           = uuid.NewString()
		userUUID         = uuid.NewString()
		mockedNamespaces = []*namespace.Namespace{
			{
				ID:       uuid.New(),
				Name:     "tenant-1",
				State:    namespace.SharedState,
				Metadata: nil,
			},
		}
	)
	type testCase struct {
		name         string
		Request      *compassv1beta1.GetNamespaceRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.NamespaceService)
		PostCheck    func(resp *compassv1beta1.GetNamespaceResponse) error
	}
	var testCases = []testCase{
		{
			name: "get namespace by its id if urn contains id",
			Request: &compassv1beta1.GetNamespaceRequest{
				Urn: mockedNamespaces[0].ID.String(),
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().GetByID(ctx, mockedNamespaces[0].ID).Return(mockedNamespaces[0], nil)
			},
			PostCheck: func(resp *compassv1beta1.GetNamespaceResponse) error {
				assert.Equal(t, resp.GetNamespace().Name, mockedNamespaces[0].Name)
				assert.Equal(t, resp.GetNamespace().Id, mockedNamespaces[0].ID.String())
				return nil
			},
		},
		{
			name: "get namespace by its name if urn contains name",
			Request: &compassv1beta1.GetNamespaceRequest{
				Urn: mockedNamespaces[0].Name,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().GetByName(ctx, mockedNamespaces[0].Name).Return(mockedNamespaces[0], nil)
			},
			PostCheck: func(resp *compassv1beta1.GetNamespaceResponse) error {
				assert.Equal(t, resp.GetNamespace().Name, mockedNamespaces[0].Name)
				assert.Equal(t, resp.GetNamespace().Id, mockedNamespaces[0].ID.String())
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockNamespaceSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockNamespaceSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, nil, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.GetNamespace(ctx, tc.Request)
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

func TestAPIServer_CreateNamespaces(t *testing.T) {
	var (
		userID           = uuid.NewString()
		userUUID         = uuid.NewString()
		mockedNamespaces = []*namespace.Namespace{
			{
				ID:    uuid.New(),
				Name:  "tenant-1",
				State: namespace.SharedState,
				Metadata: map[string]interface{}{
					"key": "value data",
				},
			},
		}
	)
	mockedNamespace0Meta, err := structpb.NewStruct(mockedNamespaces[0].Metadata)
	assert.NoError(t, err)

	type testCase struct {
		name         string
		Request      *compassv1beta1.CreateNamespaceRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.NamespaceService)
		PostCheck    func(resp *compassv1beta1.CreateNamespaceResponse) error
	}
	var testCases = []testCase{
		{
			name: "create a namespace and return its id as response",
			Request: &compassv1beta1.CreateNamespaceRequest{
				Name:     mockedNamespaces[0].Name,
				State:    mockedNamespaces[0].State.String(),
				Metadata: mockedNamespace0Meta,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().Create(ctx, mock.AnythingOfType("*namespace.Namespace")).Return(mockedNamespaces[0].ID.String(), nil)
			},
			PostCheck: func(resp *compassv1beta1.CreateNamespaceResponse) error {
				assert.NotNil(t, resp.Id)
				return nil
			},
		},
		{
			name: "throw an error if namespace already exists",
			Request: &compassv1beta1.CreateNamespaceRequest{
				Name:     mockedNamespaces[0].Name,
				State:    mockedNamespaces[0].State.String(),
				Metadata: mockedNamespace0Meta,
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().Create(ctx, mock.AnythingOfType("*namespace.Namespace")).Return("", errors.New("already exists"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockNamespaceSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockNamespaceSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, nil, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.CreateNamespace(ctx, tc.Request)
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

func TestAPIServer_UpdateNamespaces(t *testing.T) {
	var (
		userID           = uuid.NewString()
		userUUID         = uuid.NewString()
		mockedNamespaces = []*namespace.Namespace{
			{
				ID:       uuid.New(),
				Name:     "tenant-1",
				State:    namespace.SharedState,
				Metadata: nil,
			},
		}
	)
	type testCase struct {
		name         string
		Request      *compassv1beta1.UpdateNamespaceRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.NamespaceService)
		PostCheck    func(resp *compassv1beta1.UpdateNamespaceResponse) error
	}
	var testCases = []testCase{
		{
			name: "update an existing namespace state and metadata if urn is a uuid",
			Request: &compassv1beta1.UpdateNamespaceRequest{
				Urn:      mockedNamespaces[0].ID.String(),
				State:    mockedNamespaces[0].State.String(),
				Metadata: nil,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().Update(ctx, &namespace.Namespace{
					ID:       mockedNamespaces[0].ID,
					Name:     "",
					State:    mockedNamespaces[0].State,
					Metadata: nil,
				}).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockNamespaceSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockNamespaceSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, nil, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.UpdateNamespace(ctx, tc.Request)
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
