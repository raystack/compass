package grpc_interceptor

import (
	"context"
	"errors"
	"fmt"

	"github.com/odpf/columbus/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ValidateUser middleware will propagate a valid user ID as string
// within request context
// use `user.FromContext` function to get the user ID string
func ValidateUser(identityUUIDHeaderKey, identityEmailHeaderKey string, userSvc *user.Service) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", fmt.Errorf("metadata in grpc doesn't exist")
		}

		metadataValues := md.Get(identityUUIDHeaderKey)
		if len(metadataValues) < 1 {
			return nil, status.Errorf(codes.InvalidArgument, "identity header uuid is empty")
		}
		userUUID := metadataValues[0]

		var userEmail = ""
		metadataValues = md.Get(identityEmailHeaderKey)
		if len(metadataValues) > 0 {
			userEmail = metadataValues[0]
		}

		userID, err := userSvc.ValidateUser(ctx, userUUID, userEmail)
		if err != nil {
			if errors.Is(err, user.ErrNoUserInformation) {
				return nil, status.Errorf(codes.InvalidArgument, err.Error())
			}
			return nil, status.Errorf(codes.Internal, codes.Internal.String())
		}
		newCtx := user.NewContext(ctx, userID)
		return handler(newCtx, req)
	}
}
