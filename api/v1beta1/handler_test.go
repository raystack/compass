package v1beta1_test

// import (
// 	"github.com/grpc-ecosystem/grpc-gateway/runtime"
// 	"github.com/odpf/columbus/api"
// 	"github.com/odpf/columbus/asset"
// 	"github.com/odpf/columbus/discovery"
// 	"github.com/odpf/columbus/discussion"
// 	"github.com/odpf/columbus/lib/mocks"
// 	"github.com/odpf/columbus/lineage"
// 	"github.com/odpf/salt/log"
// )

// type MockDependencies struct {
// 	AssetRepository      asset.Repository
// 	DiscoveryRepository  discovery.Repository
// 	LineageRepository    lineage.Repository
// 	DiscussionRepository discussion.Repository
// }

// func setup() {
// 	mockDeps := &MockDependencies{
// 		AssetRepository:      &mocks.AssetRepository{},
// 		DiscoveryRepository:  &mocks.DiscoveryRepository{},
// 		LineageRepository:    &mocks.LineageRepository{},
// 		DiscussionRepository: &mocks.DiscussionRepository{},
// 	}
// 	handlers := api.NewHandlers(log.NewNoop(), &api.Dependencies{
// 		Logger:               log.NewNoop(),
// 		AssetRepository:      mockDeps.AssetRepository,
// 		DiscoveryRepository:  mockDeps.DiscoveryRepository,
// 		LineageRepository:    mockDeps.LineageRepository,
// 		DiscussionRepository: mockDeps.DiscussionRepository,
// 	})
// 	mux := runtime.NewServeMux()
// 	mux.RegisterService(
// 		&compassv1beta1.CompassService_ServiceDesc,
// 		handlers.GRPCHandler,
// 	)

// }
