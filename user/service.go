package user

import (
	"context"
	"fmt"
)

type contextKey1Type struct{}
type contextKey2Type struct{}

var (
	// userIDContextKey is the key used for user.FromContext and
	// user.NewContext.
	userIDContextKey = contextKey1Type(struct{}{})
	// userEmailContextKey is the key used for user.FromContext and
	// user.NewContext.
	userEmailContextKey = contextKey2Type(struct{}{})
)

// Service is a type of service that manages business process
type Service struct {
	repository Repository
}

// ValidateUser checks if user information is already in DB
// if exist in DB, return user ID, if not exist in DB, create a new one
func (s *Service) ValidateUser(ctx context.Context, email string) (string, error) {
	if email == "" {
		return "", ErrNoUserInformation
	}

	userID, err := s.repository.GetID(ctx, email)
	if err == nil {
		if userID != "" {
			return userID, nil
		}
		return "", fmt.Errorf("%w, fetched user id from DB is nil with email: %s", ErrNoUserInformation, email)
	}
	user := &User{
		Email: email,
	}
	if userID, err = s.repository.Create(ctx, user); err != nil {
		return "", err
	}
	return userID, nil
}

// NewContext returns a new context.Context that carries the provided
// user ID.
func NewContext(ctx context.Context, userID string, email string) context.Context {
	newCtx := context.WithValue(ctx, userIDContextKey, userID)
	return context.WithValue(newCtx, userEmailContextKey, email)
}

// IDFromContext returns the user ID from the context if present, and empty
// otherwise.
func IDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	h, _ := ctx.Value(userIDContextKey).(string)
	if h != "" {
		return h
	}
	return h
}

// EmailFromContext returns the user email from the context if present, and empty
// otherwise.
func EmailFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	h, _ := ctx.Value(userEmailContextKey).(string)
	if h != "" {
		return h
	}
	return h
}

// NewService initializes user service
func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}
