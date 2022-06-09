package grpc_interceptor

import (
	"context"
	"time"

	"github.com/odpf/compass/pkg/statsd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

//go:generate mockery --name=StatsDClient -r --case underscore --with-expecter --structname StatsDClient --filename statsd_monitor.go --output=./mocks
type StatsDClient interface {
	Histogram(name string, value float64) *statsd.Metric
}

func StatsD(statsdReporter StatsDClient) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if statsdReporter == nil {
			return handler(ctx, req)
		}
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		statsdReporter.Histogram("responseTime", float64(time.Since(start)/time.Millisecond)).
			Tag("method", info.FullMethod).
			Tag("status", code.String()).
			Publish()
		return resp, err
	}
}
