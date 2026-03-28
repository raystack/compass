package middleware

import (
	"connectrpc.com/connect"
)

// ChainUnaryInterceptors chains multiple unary interceptors into a single interceptor.
func ChainUnaryInterceptors(interceptors ...connect.UnaryInterceptorFunc) connect.Option {
	return connect.WithInterceptors(convertToInterceptors(interceptors)...)
}

func convertToInterceptors(funcs []connect.UnaryInterceptorFunc) []connect.Interceptor {
	interceptors := make([]connect.Interceptor, len(funcs))
	for i, f := range funcs {
		interceptors[i] = f
	}
	return interceptors
}
