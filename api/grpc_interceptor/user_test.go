package grpc_interceptor

import (
	"context"
	"errors"
	"testing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/odpf/compass/lib/mocks"
	"github.com/odpf/compass/user"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	identityUUIDHeaderKey  = "Compass-User-ID"
	identityEmailHeaderKey = "Compass-User-Email"
	userID                 = "user-id"
)

type UserTestSuite struct {
	*grpc_testing.InterceptorTestSuite
	userRepo *mocks.UserRepository
}

func TestUserSuite(t *testing.T) {
	logger := log.NewNoop()
	mockUserRepo := new(mocks.UserRepository)
	userSvc := user.NewService(logger, mockUserRepo)
	s := &UserTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &dummyService{TestServiceServer: &grpc_testing.TestPingService{T: t}},
			ServerOpts: []grpc.ServerOption{
				grpc_middleware.WithUnaryServerChain(
					ValidateUser(identityUUIDHeaderKey, identityEmailHeaderKey, userSvc)),
			},
		},
		userRepo: mockUserRepo,
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
	s.userRepo.EXPECT().GetByUUID(mock.Anything, userUUID).Return(user.User{}, customError)
	s.userRepo.EXPECT().UpsertByEmail(mock.Anything, mock.Anything).Return("", customError)

	ctx := metadata.AppendToOutgoingContext(context.Background(), identityUUIDHeaderKey, userUUID, identityEmailHeaderKey, userEmail)
	_, err := s.Client.Ping(ctx, &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.Internal, code)

	s.userRepo.AssertExpectations(s.T())
}

func (s *UserTestSuite) TestUnary_HeaderPassed() {
	userEmail := "user-email"
	userUUID := "user-uuid"
	s.userRepo.EXPECT().GetByUUID(mock.Anything, userUUID).Return(user.User{ID: userID, UUID: userUUID, Email: userEmail}, nil)

	ctx := metadata.AppendToOutgoingContext(s.SimpleCtx(), identityUUIDHeaderKey, userUUID, identityEmailHeaderKey, userEmail)
	_, err := s.Client.Ping(ctx, &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.OK, code)

	s.userRepo.AssertExpectations(s.T())
}
