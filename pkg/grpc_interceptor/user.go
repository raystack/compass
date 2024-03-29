package grpc_interceptor

import (
	"context"
	"github.com/raystack/compass/core/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UserHeaderCtx middleware will propagate a valid user ID as string
// within request context
// use `user.FromContext` function to get the user ID string
func UserHeaderCtx(IdentityHeaderKeyUUID, IdentityHeaderKeyEmail string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			userUUID  = ""
			userEmail = ""
		)

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			metadataValues := md.Get(IdentityHeaderKeyUUID)
			if len(metadataValues) > 0 {
				userUUID = metadataValues[0]
			}

			metadataValues = md.Get(IdentityHeaderKeyEmail)
			if len(metadataValues) > 0 {
				userEmail = metadataValues[0]
			}

			ctx = user.NewContext(ctx, user.User{UUID: userUUID, Email: userEmail})
		}
		return handler(ctx, req)
	}
}
