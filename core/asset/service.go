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

func (s *Service) UpsertAsset(ctx context.Context, ast *Asset, upstreams, downstreams []string) (string, error) {
	assetID, err := s.UpsertAssetWithoutLineage(ctx, ast)
	if err != nil {
		return "", err
	}

	if err := s.lineageRepository.Upsert(ctx, ast.URN, upstreams, downstreams); err != nil {
		return "", err
	}

	return assetID, nil
}

func (s *Service) UpsertAssetWithoutLineage(ctx context.Context, ast *Asset) (string, error) {
	assetID, err := s.assetRepository.Upsert(ctx, ast)
	if err != nil {
		return "", err
	}

	ast.ID = assetID
	if err := s.discoveryRepository.Upsert(ctx, *ast); err != nil {
		return "", err
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
			return Asset{}, fmt.Errorf("error when getting asset by id: %w", err)
		}
	} else {
		if ast, err = s.assetRepository.GetByURN(ctx, id); err != nil {
			return Asset{}, fmt.Errorf("error when getting asset by urn: %w", err)
		}
	}

	probes, err := s.assetRepository.GetProbes(ctx, ast.URN)
	if err != nil {
		return Asset{}, fmt.Errorf("error when getting probes: %w", err)
	}

	ast.Probes = probes

	return
}

func (s *Service) GetAssetByVersion(ctx context.Context, id string, version string) (Asset, error) {
	if isValidUUID(id) {
		return s.assetRepository.GetByVersionWithID(ctx, id, version)
	}

	return s.assetRepository.GetByVersionWithURN(ctx, id, version)
}

func (s *Service) GetAssetVersionHistory(ctx context.Context, flt Filter, id string) ([]Asset, error) {
	return s.assetRepository.GetVersionHistory(ctx, flt, id)
}

func (s *Service) AddProbe(ctx context.Context, assetURN string, probe *Probe) error {
	return s.assetRepository.AddProbe(ctx, assetURN, probe)
}

func (s *Service) GetLineage(ctx context.Context, urn string, query LineageQuery) (Lineage, error) {
	edges, err := s.lineageRepository.GetGraph(ctx, urn, query)
	if err != nil {
		return Lineage{}, fmt.Errorf("get lineage: get graph edges: %w", err)
	}

	urns := newUniqueStrings(len(edges))
	urns.add(urn)
	for _, edge := range edges {
		urns.add(edge.Source, edge.Target)
	}

	assetProbes, err := s.assetRepository.GetProbesWithFilter(ctx, ProbesFilter{
		AssetURNs: urns.list(),
		MaxRows:   1,
	})
	if err != nil {
		return Lineage{}, fmt.Errorf("get lineage: get latest probes: %w", err)
	}

	assetDataAttributes := make(map[string]Asset)
	if query.WithNodeProps {
		assetDataAttributes, err = s.assetRepository.GetAssetDataAttributes(ctx, urns.list())
		if err != nil {
			return Lineage{}, fmt.Errorf("get lineage: get asset data attributes: %w", err)
		}
	}

	return Lineage{
		Edges:     edges,
		NodeAttrs: buildNodeAttrs(assetProbes, assetDataAttributes),
	}, nil
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

func buildNodeAttrs(assetProbes map[string][]Probe, attributes map[string]Asset) map[string]NodeAttributes {
	nodeAttrs := make(map[string]NodeAttributes, len(assetProbes))
	for urn, probes := range assetProbes {
		var probesInfo ProbesInfo
		if len(probes) > 0 {
			probesInfo = ProbesInfo{Latest: probes[0]}
		}

		attrs := make(map[string]interface{})
		if val, ok := attributes[urn]; ok {
			attrs = val.Data
		}

		nodeAttrs[urn] = NodeAttributes{
			Probes:     probesInfo,
			Attributes: attrs,
		}
	}

	for urn, asset := range attributes {
		if _, ok := nodeAttrs[urn]; !ok {
			nodeAttrs[urn] = NodeAttributes{
				Attributes: asset.Data,
			}
		} else {
			nodeAttrs[urn] = NodeAttributes{
				Probes:     nodeAttrs[urn].Probes,
				Attributes: asset.Data,
			}
		}
	}

	return nodeAttrs
}
