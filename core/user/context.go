package user

import (
	"context"
)

type contextKeyType struct{}

// userContextKey is the key used for user.FromContext and
// user.NewContext.
var userContextKey = contextKeyType(struct{}{})

// NewContext returns a new context.Context that carries the provided
// user ID.
func NewContext(ctx context.Context, usr User) context.Context {
	return context.WithValue(ctx, userContextKey, usr)
}

// FromContext returns the user ID from the context if present, and empty
// otherwise.
func FromContext(ctx context.Context) User {
	if ctx == nil {
		return User{}
	}
	if u, ok := ctx.Value(userContextKey).(User); ok {
		return u
	}
	return User{}
}
