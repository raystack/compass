package grpc_interceptor

import (
	"context"
	"errors"
	"testing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/user"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	identityHeaderKey = "Columbus-User-ID"
	defaultProvider   = "shield"
)

type UserTestSuite struct {
	*grpc_testing.InterceptorTestSuite
	userRepo *mocks.UserRepository
}

func TestUserSuite(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	userSvc := user.NewService(mockUserRepo, user.Config{
		IdentityProviderDefaultName: defaultProvider,
	})
	s := &UserTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &dummyService{TestServiceServer: &grpc_testing.TestPingService{T: t}},
			ServerOpts: []grpc.ServerOption{
				grpc_middleware.WithUnaryServerChain(
					ValidateUser(identityHeaderKey, userSvc)),
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
	require.EqualError(s.T(), err, "rpc error: code = InvalidArgument desc = identity header is empty")
}

func (s *UserTestSuite) TestUnary_UserServiceError() {
	userEmail := "user-email-error"
	customError := errors.New("some error")
	s.userRepo.EXPECT().GetID(mock.Anything, userEmail).Return("", customError)
	s.userRepo.EXPECT().Create(mock.Anything, mock.Anything).Return("", customError)

	ctx := metadata.AppendToOutgoingContext(context.Background(), identityHeaderKey, userEmail)
	_, err := s.Client.Ping(ctx, &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.Internal, code)

	s.userRepo.AssertExpectations(s.T())
}

func (s *UserTestSuite) TestUnary_HeaderPassed() {
	userEmail := "user-email"
	userID := "user-id"
	s.userRepo.EXPECT().GetID(mock.Anything, userEmail).Return(userID, nil)

	ctx := metadata.AppendToOutgoingContext(s.SimpleCtx(), identityHeaderKey, userEmail)
	_, err := s.Client.Ping(ctx, &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999})
	code := status.Code(err)
	require.Equal(s.T(), codes.OK, code)

	s.userRepo.AssertExpectations(s.T())
}
