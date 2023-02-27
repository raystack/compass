package namespace

import (
	"context"
	"github.com/google/uuid"
	"github.com/odpf/salt/log"
)

type Service struct {
	storageRepo   StorageRepository
	discoveryRepo DiscoveryRepository
	logger        log.Logger
}

func NewService(logger log.Logger, storageRepo StorageRepository, discoveryRepo DiscoveryRepository) *Service {
	return &Service{
		logger:        logger,
		storageRepo:   storageRepo,
		discoveryRepo: discoveryRepo,
	}
}

func (s Service) MigrateDefault(ctx context.Context) (string, error) {
	return s.Create(ctx, DefaultNamespace)
}

func (s Service) Create(ctx context.Context, namespace *Namespace) (string, error) {
	id, err := s.storageRepo.Create(ctx, namespace)
	if err != nil {
		return "", err
	}
	if err := s.discoveryRepo.CreateNamespace(ctx, namespace); err != nil {
		return "", err
	}
	return id, nil
}

// Update can't modify a namespace ID and name
func (s Service) Update(ctx context.Context, namespace *Namespace) error {
	var existingNamespace *Namespace
	var err error
	if len(namespace.Name) > 0 {
		if existingNamespace, err = s.storageRepo.GetByName(ctx, namespace.Name); err != nil {
			return err
		}
	} else {
		if existingNamespace, err = s.storageRepo.GetByID(ctx, namespace.ID); err != nil {
			return err
		}
	}
	if len(namespace.State) > 0 {
		existingNamespace.State = namespace.State
	}
	existingNamespace.Metadata = namespace.Metadata
	return s.storageRepo.Update(ctx, existingNamespace)
}

func (s Service) GetByID(ctx context.Context, id uuid.UUID) (*Namespace, error) {
	return s.storageRepo.GetByID(ctx, id)
}

func (s Service) GetByName(ctx context.Context, name string) (*Namespace, error) {
	return s.storageRepo.GetByName(ctx, name)
}

func (s Service) List(ctx context.Context) ([]*Namespace, error) {
	return s.storageRepo.List(ctx)
}
