package star

import (
	"context"
	"github.com/raystack/compass/core/namespace"

	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/user"
)

func NewService(starRepository Repository) *Service {
	return &Service{
		starRepository: starRepository,
	}
}

type Service struct {
	starRepository Repository
}

func (s *Service) GetStarredEntitiesByUserID(ctx context.Context, flt Filter, userID string) ([]entity.Entity, error) {
	return s.starRepository.GetAllEntitiesByUserID(ctx, flt, userID)
}
func (s *Service) GetStarredEntityByUserID(ctx context.Context, userID, entityID string) (entity.Entity, error) {
	return s.starRepository.GetEntityByUserID(ctx, userID, entityID)
}
func (s *Service) GetStargazers(ctx context.Context, flt Filter, entityID string) ([]user.User, error) {
	return s.starRepository.GetStargazers(ctx, flt, entityID)
}
func (s *Service) Stars(ctx context.Context, ns *namespace.Namespace, userID, entityID string) (string, error) {
	return s.starRepository.Create(ctx, ns, userID, entityID)
}
func (s *Service) Unstars(ctx context.Context, userID, entityID string) error {
	return s.starRepository.Delete(ctx, userID, entityID)
}
