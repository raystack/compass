package interceptor

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/user"
)

// UserHeaderCtx returns a new unary interceptor that propagates a valid user ID
// from request headers within the request context.
// Use `user.FromContext` function to get the user.
func UserHeaderCtx(identityHeaderKeyUUID, identityHeaderKeyEmail string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			var (
				userUUID  = ""
				userEmail = ""
			)

			userUUID = req.Header().Get(identityHeaderKeyUUID)
			userEmail = req.Header().Get(identityHeaderKeyEmail)

			ctx = user.NewContext(ctx, user.User{UUID: userUUID, Email: userEmail})
			return next(ctx, req)
		}
	}
}
