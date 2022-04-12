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
func ValidateUser(identityHeaderKey string, userSvc *user.Service) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", fmt.Errorf("metadata in grpc doesn't exist")
		}

		metadataValues := md.Get(identityHeaderKey)
		if len(metadataValues) < 1 {
			return nil, status.Errorf(codes.InvalidArgument, "identity header is empty")
		}
		userEmail := metadataValues[0]
		userID, err := userSvc.ValidateUser(ctx, userEmail)
		if err != nil {
			if errors.Is(err, user.ErrNoUserInformation) {
				return nil, status.Errorf(codes.InvalidArgument, err.Error())
			}
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		newCtx := user.NewContext(ctx, userID)
		return handler(newCtx, req)
	}
}
