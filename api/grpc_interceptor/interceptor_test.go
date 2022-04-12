package grpc_interceptor

import (
	"context"

	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
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
	return s.TestServiceServer.Ping(ctx, ping)
}
