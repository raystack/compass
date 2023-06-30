package grpc_interceptor

import (
	"context"
	"testing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/raystack/compass/pkg/grpc_interceptor/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StatsDTestSuite struct {
	*grpc_testing.InterceptorTestSuite
	statsdClient *mocks.StatsDClient
}

func TestStatsDSuite(t *testing.T) {
	statsdClient := new(mocks.StatsDClient)

	s := &StatsDTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &dummyService{TestServiceServer: &grpc_testing.TestPingService{T: t}},
			ServerOpts: []grpc.ServerOption{
				grpc_middleware.WithUnaryServerChain(
					StatsD(statsdClient)),
			},
		},
		statsdClient: statsdClient,
	}
	suite.Run(t, s)
}

func (s *StatsDTestSuite) TestUnary_StatsDMetrics() {
	s.statsdClient.EXPECT().Histogram("responseTime", float64(0)).Return(nil).Once()
	_, err := s.Client.Ping(context.Background(), &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.OK, code)
	s.statsdClient.AssertCalled(s.T(), "Histogram", "responseTime", float64(0))
}
