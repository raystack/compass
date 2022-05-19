package handlersv1beta1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/compass/internal/server/v1beta1/mocks"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetLineageGraph(t *testing.T) {

	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	t.Run("get Lineage", func(t *testing.T) {
		t.Run("should return a graph containing the requested resource, along with it's related resources", func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
			logger := log.NewNoop()
			node := asset.Node{
				URN: "job-1",
			}
			var graph = asset.Graph{
				{Source: "job-1", Target: "table-2"},
				{Source: "table-2", Target: "table-31"},
				{Source: "table-31", Target: "dashboard-30"},
			}
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetLineage(ctx, node).Return(graph, nil)
			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.GetGraph(ctx, &compassv1beta1.GetGraphRequest{
				Urn: node.URN,
			})
			code := status.Code(err)
			if code != codes.OK {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", codes.OK, code.String())
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
			}
			if diff := cmp.Diff(got, expected, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", expected, got)
			}
		})

	})
}
