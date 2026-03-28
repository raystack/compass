package interceptor

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

// ErrorResponse returns a new unary interceptor that standardizes error formatting.
// It ensures all errors returned from handlers are proper Connect errors.
func ErrorResponse() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				return resp, ensureConnectError(err)
			}
			return resp, nil
		}
	}
}

// ensureConnectError wraps non-Connect errors as internal Connect errors.
func ensureConnectError(err error) error {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return err
	}
	return connect.NewError(connect.CodeInternal, err)
}
