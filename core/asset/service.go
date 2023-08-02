package asset

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Service struct {
	assetRepository     Repository
	discoveryRepository DiscoveryRepository
	lineageRepository   LineageRepository
	worker              Worker

	assetOpCounter metric.Int64Counter
}

//go:generate mockery --name=Worker -r --case underscore --with-expecter --structname Worker --filename worker_mock.go --output=./mocks

type Worker interface {
	EnqueueIndexAssetJob(ctx context.Context, ast Asset) error
	EnqueueDeleteAssetJob(ctx context.Context, urn string) error
	Close() error
}

type ServiceDeps struct {
	AssetRepo     Repository
	DiscoveryRepo DiscoveryRepository
	LineageRepo   LineageRepository
	Worker        Worker
}

func NewService(deps ServiceDeps) *Service {
	assetOpCounter, err := otel.Meter("github.com/goto/compass/core/asset").
		Int64Counter("compass.asset.operation")
	if err != nil {
		otel.Handle(err)
	}

	return &Service{
		assetRepository:     deps.AssetRepo,
		discoveryRepository: deps.DiscoveryRepo,
		lineageRepository:   deps.LineageRepo,
		worker:              deps.Worker,

		assetOpCounter: assetOpCounter,
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
	if err := s.worker.EnqueueIndexAssetJob(ctx, *ast); err != nil {
		return "", err
	}

	return assetID, nil
}

func (s *Service) DeleteAsset(ctx context.Context, id string) (err error) {
	defer func() {
		s.instrumentAssetOp(ctx, "DeleteAsset", id, err)
	}()

	urn := id
	if isValidUUID(id) {
		asset, err := s.assetRepository.GetByID(ctx, id)
		if err != nil {
			return err
		}

		urn = asset.URN
	}

	if err := s.assetRepository.DeleteByURN(ctx, urn); err != nil {
		return err
	}

	if err := s.worker.EnqueueDeleteAssetJob(ctx, urn); err != nil {
		return err
	}

	return s.lineageRepository.DeleteByURN(ctx, urn)
}

func (s *Service) GetAssetByID(ctx context.Context, id string) (ast Asset, err error) {
	defer func() {
		s.instrumentAssetOp(ctx, "GetAssetByID", id, err)
	}()

	if isValidUUID(id) {
		var err error
		ast, err = s.assetRepository.GetByID(ctx, id)
		if err != nil {
			return Asset{}, fmt.Errorf("get asset by id: %w", err)
		}
	} else {
		var err error
		ast, err = s.assetRepository.GetByURN(ctx, id)
		if err != nil {
			return Asset{}, fmt.Errorf("get asset by urn: %w", err)
		}
	}

	probes, err := s.assetRepository.GetProbes(ctx, ast.URN)
	if err != nil {
		return Asset{}, fmt.Errorf("get probes: %w", err)
	}

	ast.Probes = probes

	return ast, nil
}

func (s *Service) GetAssetByVersion(ctx context.Context, id, version string) (ast Asset, err error) {
	defer func() {
		s.instrumentAssetOp(ctx, "GetAssetByVersion", id, err)
	}()

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

	return Lineage{
		Edges:     edges,
		NodeAttrs: buildNodeAttrs(assetProbes),
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

func (s *Service) GroupAssets(ctx context.Context, cfg GroupConfig) (results []GroupResult, err error) {
	return s.discoveryRepository.GroupAssets(ctx, cfg)
}

func (s *Service) SuggestAssets(ctx context.Context, cfg SearchConfig) (suggestions []string, err error) {
	return s.discoveryRepository.Suggest(ctx, cfg)
}

func (s *Service) instrumentAssetOp(ctx context.Context, op, id string, err error) {
	identifier := "URN"
	if isValidUUID(id) {
		identifier = "ID"
	}

	s.assetOpCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("compass.asset_operation", op),
		attribute.String("asset.identifier", identifier),
		attribute.Bool("operation.success", err == nil),
	))
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func buildNodeAttrs(assetProbes map[string][]Probe) map[string]NodeAttributes {
	nodeAttrs := make(map[string]NodeAttributes, len(assetProbes))
	for urn, probes := range assetProbes {
		if len(probes) == 0 {
			continue
		}

		nodeAttrs[urn] = NodeAttributes{
			Probes: ProbesInfo{Latest: probes[0]},
		}
	}

	return nodeAttrs
}
