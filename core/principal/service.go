package principal

import (
	"context"
	"errors"

	"github.com/raystack/compass/core/namespace"
)

// Service is a type of service that manages principal business logic.
type Service struct {
	repository Repository
}

// ValidatePrincipal checks if the principal's subject is already in DB.
// If it exists, return the principal ID; if not, create a new one.
func (s *Service) ValidatePrincipal(ctx context.Context, ns *namespace.Namespace, subject, name, pType string) (string, error) {
	if subject == "" {
		return "", ErrNoPrincipalInformation
	}

	p, err := s.repository.GetBySubject(ctx, subject)
	if err == nil {
		if p.ID != "" {
			return p.ID, nil
		}
		return "", errors.New("fetched principal ID from DB is empty")
	}
	if !errors.As(err, &NotFoundError{}) {
		return "", err
	}

	if pType == "" {
		pType = "user"
	}

	id, err := s.repository.Upsert(ctx, ns, &Principal{
		Subject: subject,
		Name:    name,
		Type:    pType,
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// NewService initializes principal service.
func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}
