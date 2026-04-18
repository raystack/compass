package principal

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname PrincipalRepository --filename principal_repository.go --output=./mocks

import (
	"context"
	"time"

	"github.com/raystack/compass/core/namespace"
)

// Principal represents an identity that can act in the system.
// It may be a human user, an AI agent, or a service account.
type Principal struct {
	ID        string         `json:"id" db:"id"`
	Type      string         `json:"type" db:"type"`           // "user", "agent", "service"
	Name      string         `json:"name" db:"name"`
	Subject   string         `json:"subject" db:"subject"`     // JWT sub claim, unique external identity
	Metadata  map[string]any `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time      `json:"-" db:"created_at"`
	UpdatedAt time.Time      `json:"-" db:"updated_at"`
}

// Validate validates a principal is valid or not.
func (p *Principal) Validate() error {
	if p == nil {
		return ErrNoPrincipalInformation
	}

	if p.Subject == "" {
		return InvalidError{Subject: p.Subject}
	}

	return nil
}

// Repository contains interface of supported methods.
type Repository interface {
	Create(ctx context.Context, ns *namespace.Namespace, p *Principal) (string, error)
	GetBySubject(ctx context.Context, subject string) (Principal, error)
	Upsert(ctx context.Context, ns *namespace.Namespace, p *Principal) (string, error)
}
