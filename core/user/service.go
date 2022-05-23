package user

import (
	"context"
	"errors"

	"github.com/odpf/salt/log"
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

// NewService initializes user service
func NewService(logger log.Logger, repository Repository) *Service {
	return &Service{
		repository: repository,
		logger:     logger,
	}
}
