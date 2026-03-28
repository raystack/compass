package middleware

import (
	"context"
	"time"

	"connectrpc.com/connect"
	log "github.com/raystack/salt/observability/logger"
)

// Logger returns a new unary interceptor that logs request details.
func Logger(logger log.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			procedure := req.Spec().Procedure

			resp, err := next(ctx, req)

			duration := time.Since(start)
			fields := []interface{}{
				"procedure", procedure,
				"duration_ms", duration.Milliseconds(),
				"peer_addr", req.Peer().Addr,
			}

			if err != nil {
				fields = append(fields, "error", err.Error())
				connectErr, ok := err.(*connect.Error)
				if ok {
					fields = append(fields, "code", connectErr.Code().String())
				}
				logger.Error("request failed", fields...)
			} else {
				logger.Debug("request completed", fields...)
			}

			return resp, err
		}
	}
}
