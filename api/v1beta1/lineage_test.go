package v1beta1_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/odpf/columbus/api"
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetLineageGraph(t *testing.T) {

	t.Run("get Lineage", func(t *testing.T) {
		t.Run("should return a graph containing the requested resource, along with it's related resources", func(t *testing.T) {
			ctx := context.Background()
			logger := log.NewNoop()
			node := lineage.Node{
				URN: "job-1",
			}
			var graph = lineage.Graph{
				{Source: "job-1", Target: "table-2"},
				{Source: "table-2", Target: "table-31"},
				{Source: "table-31", Target: "dashboard-30"},
			}
			lr := new(mocks.LineageRepository)
			lr.EXPECT().GetGraph(ctx, node).Return(graph, nil)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				LineageRepository: lr,
			})

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
