package user

import (
	"context"
	"errors"

	"github.com/goto/compass/pkg/statsd"
	"github.com/goto/salt/log"
)

// Service is a type of service that manages business process
type Service struct {
	statsdReporter *statsd.Reporter
	repository     Repository
	logger         log.Logger
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
			s.statsdReporter.Incr("user_stats").
				Tag("info", "existing").
				Publish()
			return usr.ID, nil
		}
		s.statsdReporter.Incr("user_stats").
			Tag("info", "error").
			Publish()
		err := errors.New("fetched user uuid from DB is empty")
		s.logger.Error(err.Error())
		return "", err
	}

	uid, err := s.repository.UpsertByEmail(ctx, &User{
		UUID:  uuid,
		Email: email,
	})
	if err != nil {
		s.statsdReporter.Incr("user_stats").
			Tag("info", "error").
			Publish()
		s.logger.Error("error when UpsertByEmail in ValidateUser service", "err", err.Error())
		return "", err
	}
	s.statsdReporter.Incr("user_stats").
		Tag("info", "new").
		Publish()
	return uid, nil
}

// NewService initializes user service
func NewService(logger log.Logger, repository Repository, opts ...func(*Service)) *Service {
	s := &Service{
		repository: repository,
		logger:     logger,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.statsdReporter == nil {
		s.statsdReporter = &statsd.Reporter{}
	}

	return s
}

func ServiceWithStatsDReporter(statsdReporter *statsd.Reporter) func(*Service) {
	return func(s *Service) {
		s.statsdReporter = statsdReporter
	}
}
