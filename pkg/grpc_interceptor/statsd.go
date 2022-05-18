package grpc_interceptor

import (
	"context"
	"time"

	"github.com/odpf/compass/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func StatsD(mm *metrics.StatsdMonitor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if mm == nil {
			return handler(ctx, req)
		}
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		mm.ResponseTimeGRPC(info.FullMethod, int64(time.Since(start)/time.Millisecond))
		mm.ResponseStatusGRPC(info.FullMethod, code.String())
		return resp, err
	}
}
