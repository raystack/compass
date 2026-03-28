package interceptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	log "github.com/raystack/salt/observability/logger"
)

// ErrorResponse returns a new unary interceptor that standardizes error formatting.
// It ensures all errors returned from handlers are proper Connect errors.
// Non-Connect errors are sanitized to prevent leaking internal details.
func ErrorResponse(logger log.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				return resp, ensureConnectError(logger, err)
			}
			return resp, nil
		}
	}
}

// ensureConnectError wraps non-Connect errors as sanitized internal Connect errors.
func ensureConnectError(logger log.Logger, err error) error {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return err
	}
	ref := time.Now().Unix()
	logger.Error(err.Error(), "ref", ref)
	return connect.NewError(connect.CodeInternal, fmt.Errorf(
		"%s - ref (%d)",
		http.StatusText(http.StatusInternalServerError),
		ref,
	))
}
