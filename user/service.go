package user

import "context"

// Service is a type of service that manages business process
type Service struct {
	repository Repository
}

// Create handles create business operation for user
func (s *Service) GetID(ctx context.Context, email string) (string, error) {
	return s.repository.GetID(ctx, email)
}

// Create handles create business operation for user
func (s *Service) Create(ctx context.Context, user *User) (string, error) {
	return s.repository.Create(ctx, user)
}

// NewService initializes user service
func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}
