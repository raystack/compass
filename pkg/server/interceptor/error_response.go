package interceptor

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorResponse returns a new unary interceptor that standardizes error formatting.
// It converts gRPC status errors to Connect errors for compatibility.
func ErrorResponse() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				return resp, convertError(err)
			}
			return resp, nil
		}
	}
}

// convertError converts gRPC status errors to Connect errors.
func convertError(err error) error {
	// If already a Connect error, return as-is
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return err
	}

	// Convert gRPC status errors to Connect errors
	if st, ok := status.FromError(err); ok {
		return connect.NewError(grpcCodeToConnectCode(st.Code()), errors.New(st.Message()))
	}

	// Return as internal error
	return connect.NewError(connect.CodeInternal, err)
}

// grpcCodeToConnectCode maps gRPC codes to Connect codes.
func grpcCodeToConnectCode(code codes.Code) connect.Code {
	switch code {
	case codes.OK:
		return 0
	case codes.Canceled:
		return connect.CodeCanceled
	case codes.Unknown:
		return connect.CodeUnknown
	case codes.InvalidArgument:
		return connect.CodeInvalidArgument
	case codes.DeadlineExceeded:
		return connect.CodeDeadlineExceeded
	case codes.NotFound:
		return connect.CodeNotFound
	case codes.AlreadyExists:
		return connect.CodeAlreadyExists
	case codes.PermissionDenied:
		return connect.CodePermissionDenied
	case codes.ResourceExhausted:
		return connect.CodeResourceExhausted
	case codes.FailedPrecondition:
		return connect.CodeFailedPrecondition
	case codes.Aborted:
		return connect.CodeAborted
	case codes.OutOfRange:
		return connect.CodeOutOfRange
	case codes.Unimplemented:
		return connect.CodeUnimplemented
	case codes.Internal:
		return connect.CodeInternal
	case codes.Unavailable:
		return connect.CodeUnavailable
	case codes.DataLoss:
		return connect.CodeDataLoss
	case codes.Unauthenticated:
		return connect.CodeUnauthenticated
	default:
		return connect.CodeUnknown
	}
}
