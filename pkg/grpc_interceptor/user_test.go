package grpc_interceptor

import (
	"context"
	"errors"
	"testing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/odpf/compass/internal/server/v1beta1/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	IdentityHeaderUUIDKey  = "Compass-User-ID"
	IdentityHeaderEmailKey = "Compass-User-Email"
	userID                 = "user-id"
)

type UserTestSuite struct {
	*grpc_testing.InterceptorTestSuite
	userSvc *mocks.UserService
}

func TestUserSuite(t *testing.T) {
	mockUserSvc := new(mocks.UserService)
	s := &UserTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &dummyService{TestServiceServer: &grpc_testing.TestPingService{T: t}},
			ServerOpts: []grpc.ServerOption{
				grpc_middleware.WithUnaryServerChain(
					ValidateUser(IdentityHeaderUUIDKey, IdentityHeaderEmailKey, mockUserSvc)),
			},
		},
		userSvc: mockUserSvc,
	}
	suite.Run(t, s)
}

func (s *UserTestSuite) TestUnary_IdentityHeaderNotPresent() {
	_, err := s.Client.Ping(s.SimpleCtx(), &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.InvalidArgument, code)
	require.EqualError(s.T(), err, "rpc error: code = InvalidArgument desc = identity header uuid is empty")
}

func (s *UserTestSuite) TestUnary_UserServiceError() {
	userEmail := "user-email-error"
	userUUID := "user-uuid-error"
	customError := errors.New("some error")
	s.userSvc.EXPECT().ValidateUser(mock.Anything, userUUID, userEmail).Return("", customError)

	ctx := metadata.AppendToOutgoingContext(context.Background(), IdentityHeaderUUIDKey, userUUID, IdentityHeaderEmailKey, userEmail)
	_, err := s.Client.Ping(ctx, &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.Internal, code)

	s.userSvc.AssertExpectations(s.T())
}

func (s *UserTestSuite) TestUnary_HeaderPassed() {
	userEmail := "user-email"
	userUUID := "user-uuid"
	s.userSvc.EXPECT().ValidateUser(mock.Anything, userUUID, userEmail).Return(userID, nil)

	ctx := metadata.AppendToOutgoingContext(s.SimpleCtx(), IdentityHeaderUUIDKey, userUUID, IdentityHeaderEmailKey, userEmail)
	_, err := s.Client.Ping(ctx, &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.OK, code)

	s.userSvc.AssertExpectations(s.T())
}
