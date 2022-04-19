package grpc_interceptor

import (
	"context"
	"testing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/odpf/compass/lib/mocks"
	"github.com/odpf/compass/metrics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	statsdPrefix     = "compassApi"
	metricsSeparator = "."
)

type StatsDTestSuite struct {
	*grpc_testing.InterceptorTestSuite
	statsdClient *mocks.StatsdClient
}

func TestStatsDSuite(t *testing.T) {
	statsdClient := new(mocks.StatsdClient)

	monitor := metrics.NewStatsdMonitor(statsdClient, statsdPrefix, metricsSeparator)
	s := &StatsDTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &dummyService{TestServiceServer: &grpc_testing.TestPingService{T: t}},
			ServerOpts: []grpc.ServerOption{
				grpc_middleware.WithUnaryServerChain(
					StatsD(monitor)),
			},
		},
		statsdClient: statsdClient,
	}
	suite.Run(t, s)
}

func (s *StatsDTestSuite) TestUnary_StatsDMetrics() {
	s.statsdClient.EXPECT().Increment("compassApi.responseStatusCode,statusCode=OK,method=/mwitkow.testproto.TestService/Ping").Once()
	s.statsdClient.EXPECT().Timing("compassApi.responseTime,method=/mwitkow.testproto.TestService/Ping", int64(0)).Once()
	_, err := s.Client.Ping(context.Background(), &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.OK, code)
	s.statsdClient.AssertExpectations(s.T())
}
