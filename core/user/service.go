package user

import (
	"context"
	"errors"

	"github.com/odpf/salt/log"
)

type contextKeyType struct{}

var (
	// userContextKey is the key used for user.FromContext and
	// user.NewContext.
	userContextKey = contextKeyType(struct{}{})
)

// Service is a type of service that manages business process
type Service struct {
	repository Repository
	logger     log.Logger
}

// ValidateUser checks if user uuid is already in DB
// if exist in DB, return user ID, if not exist in DB, create a new one
func (s *Service) ValidateUser(ctx context.Context, uuid, email string) (string, error) {
	if uuid == "" {
		return "", ErrNoUserInformation
	}

	usr, err := s.repository.GetByUUID(ctx, uuid)
	if err == nil {
		if usr.ID != "" {
			return usr.ID, nil
		}
		err := errors.New("fetched user uuid from DB is empty")
		s.logger.Error(err.Error())
		return "", err
	}

	uid, err := s.repository.UpsertByEmail(ctx, &User{
		UUID:  uuid,
		Email: email,
	})
	if err != nil {
		s.logger.Error("error when UpsertByEmail in ValidateUser service", "err", err.Error())
		return "", err
	}
	return uid, nil
}

// NewContext returns a new context.Context that carries the provided
// user ID.
func NewContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userContextKey, userID)
}

// FromContext returns the user ID from the context if present, and empty
// otherwise.
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	h, _ := ctx.Value(userContextKey).(string)
	if h != "" {
		return h
	}
	return h
}

// NewService initializes user service
func NewService(logger log.Logger, repository Repository) *Service {
	return &Service{
		repository: repository,
		logger:     logger,
	}
}
