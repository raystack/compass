package grpc_interceptor

import (
	"context"

	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/odpf/compass/core/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type dummyService struct {
	pb_testproto.TestServiceServer
}

func (s *dummyService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	if ping.Value == "panic" {
		panic("very bad thing happened")
	}
	if ping.Value == "nilpanic" {
		panic(nil)
	}
	if ping.Value == "testuser" {
		usr := user.FromContext(ctx)
		if usr.UUID == "" {
			return nil, status.Error(codes.InvalidArgument, "uuid not found")
		}
	}
	return s.TestServiceServer.Ping(ctx, ping)
}
