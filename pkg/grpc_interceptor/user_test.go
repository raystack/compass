package grpc_interceptor

import (
	"context"
	"testing"

	"github.com/raystack/compass/core/user"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	IdentityHeaderKeyUUID  = "Compass-User-ID"
	IdentityHeaderKeyEmail = "Compass-User-Email"
)

func TestUserHeaderCtx(t *testing.T) {
	interceptor := UserHeaderCtx(IdentityHeaderKeyUUID, IdentityHeaderKeyEmail)

	// handler mimics a gRPC service that requires user UUID in context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		u := user.FromContext(ctx)
		if u.UUID == "" {
			return nil, status.Error(codes.InvalidArgument, "uuid not found")
		}
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Test"}

	t.Run("IdentityHeaderNotPresent", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
		_, err := interceptor(ctx, nil, info, handler)
		code := status.Code(err)
		require.Equal(t, codes.InvalidArgument, code)
		require.EqualError(t, err, "rpc error: code = InvalidArgument desc = uuid not found")
	})

	t.Run("HeaderPresentAndEmpty", func(t *testing.T) {
		md := metadata.Pairs(IdentityHeaderKeyUUID, "", IdentityHeaderKeyEmail, "")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err := interceptor(ctx, nil, info, handler)
		code := status.Code(err)
		require.Equal(t, codes.InvalidArgument, code)
		require.EqualError(t, err, "rpc error: code = InvalidArgument desc = uuid not found")
	})

	t.Run("HeaderPresentAndPassed", func(t *testing.T) {
		userEmail := "user-email"
		userUUID := "user-uuid"

		md := metadata.Pairs(IdentityHeaderKeyUUID, userUUID, IdentityHeaderKeyEmail, userEmail)
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err := interceptor(ctx, nil, info, handler)
		code := status.Code(err)
		require.Equal(t, codes.OK, code)
	})
}
