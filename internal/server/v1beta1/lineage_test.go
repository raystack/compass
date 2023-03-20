package handlersv1beta1

import (
	"context"
	"github.com/odpf/compass/core/namespace"
	"github.com/odpf/compass/pkg/grpc_interceptor"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetLineageGraph(t *testing.T) {
	// TODO[2022-10-13|@sudo-suhas]: Add comprehensive tests
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
		ns       = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "tenant",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)
	ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
	ctx = grpc_interceptor.BuildContextWithNamespace(ctx, ns)
	t.Run("get Lineage", func(t *testing.T) {
		t.Run("should return a graph containing the requested resource, along with it's related resources", func(t *testing.T) {
			logger := log.NewNoop()
			nodeURN := "job-1"
			level := 8
			direction := asset.LineageDirectionUpstream
			ts := time.Unix(1665659885, 0)
			tspb := timestamppb.New(ts)

			lineage := asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "job-1", Target: "table-2"},
					{Source: "table-2", Target: "table-31"},
					{Source: "table-31", Target: "dashboard-30"},
				},
				NodeAttrs: map[string]asset.NodeAttributes{
					"job-1": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "SUCCESS", Timestamp: ts, CreatedAt: ts},
						},
					},
					"table-2": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "FAILED", Timestamp: ts, CreatedAt: ts},
						},
					},
				},
			}
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)
			defer mockNamespaceSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetLineage(ctx, nodeURN, asset.LineageQuery{Level: level, Direction: direction}).Return(lineage, nil)
			mockUserSvc.EXPECT().ValidateUser(ctx, ns, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, mockSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.GetGraph(ctx, &compassv1beta1.GetGraphRequest{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
			})
			code := status.Code(err)
			if code != codes.OK {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.OK, code.String())
				return
			}

			expected := &compassv1beta1.GetGraphResponse{
				Data: []*compassv1beta1.LineageEdge{
					{
						Source: "job-1",
						Target: "table-2",
					},
					{
						Source: "table-2",
						Target: "table-31",
					},
					{
						Source: "table-31",
						Target: "dashboard-30",
					},
				},
				NodeAttrs: map[string]*compassv1beta1.GetGraphResponse_NodeAttributes{
					"job-1": {
						Probes: &compassv1beta1.GetGraphResponse_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "SUCCESS", Timestamp: tspb, CreatedAt: tspb},
						},
					},
					"table-2": {
						Probes: &compassv1beta1.GetGraphResponse_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "FAILED", Timestamp: tspb, CreatedAt: tspb},
						},
					},
				},
			}
			if diff := cmp.Diff(got, expected, protocmp.Transform()); diff != "" {
				t.Errorf("expected: %+v\ngot: %+v\ndiff: %s\n", expected, got, diff)
			}
		})

	})
}
