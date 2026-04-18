package middleware

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/principal"
)

// PrincipalHeaderCtx returns a new unary interceptor that propagates a valid principal
// from request headers within the request context.
// Use `principal.FromContext` function to get the principal.
func PrincipalHeaderCtx(identityHeaderKeyUUID, identityHeaderKeyEmail string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			subject := req.Header().Get(identityHeaderKeyUUID)
			name := req.Header().Get(identityHeaderKeyEmail)
			pType := req.Header().Get("X-Principal-Type")

			if pType == "" {
				pType = "user"
			}

			ctx = principal.NewContext(ctx, principal.Principal{
				Subject: subject,
				Name:    name,
				Type:    pType,
			})
			return next(ctx, req)
		}
	}
}
