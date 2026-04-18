package principal

import (
	"context"
)

type contextKeyType struct{}

var (
	// principalContextKey is the key used for principal.FromContext and
	// principal.NewContext.
	principalContextKey = contextKeyType(struct{}{})
)

// NewContext returns a new context.Context that carries the provided
// principal.
func NewContext(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, p)
}

// FromContext returns the principal from the context if present, and empty
// otherwise.
func FromContext(ctx context.Context) Principal {
	if ctx == nil {
		return Principal{}
	}
	if p, ok := ctx.Value(principalContextKey).(Principal); ok {
		return p
	}
	return Principal{}
}
