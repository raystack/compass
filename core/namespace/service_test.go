package namespace_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/odpf/compass/core/namespace"
	"github.com/odpf/compass/core/namespace/mocks"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_Create(t *testing.T) {
	ctx := context.TODO()
	mockedNamespace := &namespace.Namespace{
		ID:       uuid.New(),
		Name:     "tenant-1",
		State:    namespace.SharedState,
		Metadata: nil,
	}
	type request struct {
		ns *namespace.Namespace
	}
	type response struct {
		id string
	}
	type testCase struct {
		name      string
		Request   request
		ExpectErr bool
		Setup     func(context.Context, *mocks.NamespaceStorageRepository, *mocks.NamespaceDiscoveryRepository)
		PostCheck func(resp response) error
	}
	var testCases = []testCase{
		{
			name:      "create a new namespace in storage repo and then in discovery repo",
			Request:   request{mockedNamespace},
			ExpectErr: false,
			Setup: func(ctx context.Context, sr *mocks.NamespaceStorageRepository, dr *mocks.NamespaceDiscoveryRepository) {
				sr.EXPECT().Create(ctx, mockedNamespace).Return(mockedNamespace.ID.String(), nil)
				dr.EXPECT().CreateNamespace(ctx, mockedNamespace).Return(nil)
			},
			PostCheck: func(resp response) error {
				assert.NotNil(t, resp.id)
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			logger := log.NewNoop()
			mockStorageRepo := new(mocks.NamespaceStorageRepository)
			mockDiscoveryRepo := new(mocks.NamespaceDiscoveryRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockStorageRepo, mockDiscoveryRepo)
			}
			defer mockStorageRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)

			service := namespace.NewService(logger, mockStorageRepo, mockDiscoveryRepo)
			got, err := service.Create(ctx, tc.Request.ns)
			if !tc.ExpectErr && err != nil {
				t.Errorf("expected handler to not return err but got %s", err.Error())
				return
			}
			if tc.ExpectErr && err == nil {
				t.Errorf("expected handler to return err but got no error")
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(response{got}); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	ctx := context.TODO()
	mockedNamespace := &namespace.Namespace{
		ID:       uuid.New(),
		Name:     "tenant-1",
		State:    namespace.SharedState,
		Metadata: map[string]interface{}{},
	}
	type request struct {
		ns *namespace.Namespace
	}
	type testCase struct {
		name      string
		Request   request
		ExpectErr bool
		Setup     func(context.Context, *mocks.NamespaceStorageRepository, *mocks.NamespaceDiscoveryRepository)
	}
	var testCases = []testCase{
		{
			name: "throw an error if namespace doesn't exists already",
			Request: request{&namespace.Namespace{
				ID:       mockedNamespace.ID,
				Name:     "",
				State:    mockedNamespace.State,
				Metadata: mockedNamespace.Metadata,
			}},
			ExpectErr: true,
			Setup: func(ctx context.Context, sr *mocks.NamespaceStorageRepository, dr *mocks.NamespaceDiscoveryRepository) {
				sr.EXPECT().GetByID(ctx, mockedNamespace.ID).Return(nil, errors.New("doesn't exist"))
			},
		},
		{
			name: "update an existing namespace in storage repo by its id",
			Request: request{&namespace.Namespace{
				ID:    mockedNamespace.ID,
				Name:  "",
				State: mockedNamespace.State,
				Metadata: map[string]interface{}{
					"hello": "world",
				},
			}},
			ExpectErr: false,
			Setup: func(ctx context.Context, sr *mocks.NamespaceStorageRepository, dr *mocks.NamespaceDiscoveryRepository) {
				sr.EXPECT().GetByID(ctx, mockedNamespace.ID).Return(&namespace.Namespace{
					ID:       mockedNamespace.ID,
					Name:     mockedNamespace.Name,
					State:    mockedNamespace.State,
					Metadata: mockedNamespace.Metadata,
				}, nil)
				sr.EXPECT().Update(ctx, &namespace.Namespace{
					ID:    mockedNamespace.ID,
					Name:  mockedNamespace.Name,
					State: mockedNamespace.State,
					Metadata: map[string]interface{}{
						"hello": "world",
					},
				}).Return(nil)
			},
		},
		{
			name: "update an existing namespace in storage repo by its name",
			Request: request{&namespace.Namespace{
				Name:  mockedNamespace.Name,
				State: mockedNamespace.State,
				Metadata: map[string]interface{}{
					"hello": "world",
				},
			}},
			ExpectErr: false,
			Setup: func(ctx context.Context, sr *mocks.NamespaceStorageRepository, dr *mocks.NamespaceDiscoveryRepository) {
				sr.EXPECT().GetByName(ctx, mockedNamespace.Name).Return(&namespace.Namespace{
					ID:       mockedNamespace.ID,
					Name:     mockedNamespace.Name,
					State:    mockedNamespace.State,
					Metadata: mockedNamespace.Metadata,
				}, nil)
				sr.EXPECT().Update(ctx, &namespace.Namespace{
					ID:    mockedNamespace.ID,
					Name:  mockedNamespace.Name,
					State: mockedNamespace.State,
					Metadata: map[string]interface{}{
						"hello": "world",
					},
				}).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			logger := log.NewNoop()
			mockStorageRepo := new(mocks.NamespaceStorageRepository)
			mockDiscoveryRepo := new(mocks.NamespaceDiscoveryRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockStorageRepo, mockDiscoveryRepo)
			}
			defer mockStorageRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)

			service := namespace.NewService(logger, mockStorageRepo, mockDiscoveryRepo)
			err := service.Update(ctx, tc.Request.ns)
			if !tc.ExpectErr && err != nil {
				t.Errorf("expected handler to not return err but got %s", err.Error())
				return
			}
			if tc.ExpectErr && err == nil {
				t.Errorf("expected handler to return err but got no error")
				return
			}
		})
	}
}
