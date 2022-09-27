package asset

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	assetRepository     Repository
	discoveryRepository DiscoveryRepository
	lineageRepository   LineageRepository
}

func NewService(assetRepository Repository, discoveryRepository DiscoveryRepository, lineageRepository LineageRepository) *Service {
	return &Service{
		assetRepository:     assetRepository,
		discoveryRepository: discoveryRepository,
		lineageRepository:   lineageRepository,
	}
}

func (s *Service) GetAllAssets(ctx context.Context, flt Filter, withTotal bool) ([]Asset, uint32, error) {
	var totalCount uint32 = 0
	assets, err := s.assetRepository.GetAll(ctx, flt)
	if err != nil {
		return nil, totalCount, err
	}

	if withTotal {
		total, err := s.assetRepository.GetCount(ctx, flt)
		if err != nil {
			return nil, totalCount, err
		}
		totalCount = uint32(total)
	}
	return assets, totalCount, nil
}

func (s *Service) UpsertPatchAsset(ctx context.Context, ast *Asset, upstreams, downstreams []LineageNode) (string, error) {
	var assetID string
	var err error

	assetID, err = s.assetRepository.Upsert(ctx, ast)
	if err != nil {
		return assetID, err
	}

	ast.ID = assetID
	if err := s.discoveryRepository.Upsert(ctx, *ast); err != nil {
		return assetID, err
	}

	node := LineageNode{
		URN:     ast.URN,
		Type:    ast.Type,
		Service: ast.Service,
	}

	if err := s.lineageRepository.Upsert(ctx, node, upstreams, downstreams); err != nil {
		return assetID, err
	}

	return assetID, nil
}

func (s *Service) DeleteAsset(ctx context.Context, id string) error {
	if isValidUUID(id) {
		if err := s.assetRepository.DeleteByID(ctx, id); err != nil {
			return err
		}

		return s.discoveryRepository.DeleteByID(ctx, id)
	}

	if err := s.assetRepository.DeleteByURN(ctx, id); err != nil {
		return err
	}

	return s.discoveryRepository.DeleteByURN(ctx, id)
}

func (s *Service) GetAssetByID(ctx context.Context, id string) (ast Asset, err error) {
	if isValidUUID(id) {
		if ast, err = s.assetRepository.GetByID(ctx, id); err != nil {
			return ast, fmt.Errorf("error when getting asset by id: %w", err)
		}
	} else {
		if ast, err = s.assetRepository.GetByURN(ctx, id); err != nil {
			return ast, fmt.Errorf("error when getting asset by urn: %w", err)
		}
	}

	probes, err := s.assetRepository.GetProbes(ctx, ast.URN)
	if err != nil {
		return ast, fmt.Errorf("error when getting probes: %w", err)
	}

	ast.Probes = probes

	return
}

func (s *Service) GetAssetByVersion(ctx context.Context, id string, version string) (Asset, error) {
	return s.assetRepository.GetByVersion(ctx, id, version)
}

func (s *Service) GetAssetVersionHistory(ctx context.Context, flt Filter, id string) ([]Asset, error) {
	return s.assetRepository.GetVersionHistory(ctx, flt, id)
}

func (s *Service) AddProbe(ctx context.Context, assetURN string, probe *Probe) error {
	return s.assetRepository.AddProbe(ctx, assetURN, probe)
}

func (s *Service) GetLineage(ctx context.Context, node LineageNode, query LineageQuery) (LineageGraph, error) {
	return s.lineageRepository.GetGraph(ctx, node, query)
}

func (s *Service) GetTypes(ctx context.Context, flt Filter) (map[Type]int, error) {
	result, err := s.assetRepository.GetTypes(ctx, flt)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) SearchAssets(ctx context.Context, cfg SearchConfig) (results []SearchResult, err error) {
	return s.discoveryRepository.Search(ctx, cfg)
}
func (s *Service) SuggestAssets(ctx context.Context, cfg SearchConfig) (suggestions []string, err error) {
	return s.discoveryRepository.Suggest(ctx, cfg)
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
