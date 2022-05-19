package grpc_interceptor

import (
	"context"
	"testing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	IdentityHeaderKeyUUID  = "Compass-User-ID"
	IdentityHeaderKeyEmail = "Compass-User-Email"
)

type UserTestSuite struct {
	*grpc_testing.InterceptorTestSuite
}

func TestUserSuite(t *testing.T) {
	s := &UserTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &dummyService{TestServiceServer: &grpc_testing.TestPingService{T: t}},
			ServerOpts: []grpc.ServerOption{
				grpc_middleware.WithUnaryServerChain(
					UserHeaderCtx(IdentityHeaderKeyUUID, IdentityHeaderKeyEmail)),
			},
		},
	}
	suite.Run(t, s)
}

func (s *UserTestSuite) TestUnary_IdentityHeaderNotPresent() {
	_, err := s.Client.Ping(s.SimpleCtx(), &pb_testproto.PingRequest{Value: "testuser", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.InvalidArgument, code)
	require.EqualError(s.T(), err, "rpc error: code = InvalidArgument desc = uuid not found")
}

func (s *UserTestSuite) TestUnary_HeaderPresentAndEmpty() {
	ctx := metadata.AppendToOutgoingContext(context.Background(), IdentityHeaderKeyUUID, "", IdentityHeaderKeyEmail, "")
	_, err := s.Client.Ping(ctx, &pb_testproto.PingRequest{Value: "testuser", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.InvalidArgument, code)
	require.EqualError(s.T(), err, "rpc error: code = InvalidArgument desc = uuid not found")
}

func (s *UserTestSuite) TestUnary_HeaderPresentAndPassed() {
	userEmail := "user-email"
	userUUID := "user-uuid"

	ctx := metadata.AppendToOutgoingContext(s.SimpleCtx(), IdentityHeaderKeyUUID, userUUID, IdentityHeaderKeyEmail, userEmail)
	_, err := s.Client.Ping(ctx, &pb_testproto.PingRequest{Value: "testuser", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.OK, code)
}
