package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/odpf/compass/core/namespace"
	"io"
	"strings"

	"github.com/odpf/compass/core/asset"
)

// DiscoveryRepository implements discovery.Repository
// with elasticsearch as the backing store.
type DiscoveryRepository struct {
	cli     *Client
	refresh string
}

type AssetModel struct {
	asset.Asset
	NamespaceID string `json:"namespace_id"`
}

func NewDiscoveryRepository(cli *Client, opts ...func(*DiscoveryRepository)) *DiscoveryRepository {
	repo := &DiscoveryRepository{
		cli:     cli,
		refresh: "false",
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

// WithInstantRefresh refresh the affected shards to make insert operations visible to search instantly
func WithInstantRefresh() func(*DiscoveryRepository) {
	return func(repository *DiscoveryRepository) {
		repository.refresh = "true"
	}
}

func (repo *DiscoveryRepository) CreateNamespace(ctx context.Context, ns *namespace.Namespace) error {
	// check if index exists
	if exists, err := repo.cli.IndexExists(ctx, ns); err != nil {
		return err
	} else if !exists {
		// doesn't exist yet, create one
		if err := repo.cli.CreateIndex(ctx, BuildIndexNameFromNamespace(ns), DefaultShardCountPerIndex); err != nil {
			return err
		}
	}
	// create alias over index, doesn't matter if its shared or dedicated
	return repo.cli.CreateIdxAlias(ctx, ns)
}

func (repo *DiscoveryRepository) Upsert(ctx context.Context, ns *namespace.Namespace, ast *asset.Asset) error {
	if ast.ID == "" {
		return asset.ErrEmptyID
	}
	if !ast.Type.IsValid() {
		return asset.ErrUnknownType
	}

	body, err := repo.createUpsertBody(ns, ast)
	if err != nil {
		return asset.DiscoveryError{Err: fmt.Errorf("error serialising payload: %w", err)}
	}
	res, err := repo.cli.client.Bulk(
		body,
		repo.cli.client.Bulk.WithContext(ctx),
		repo.cli.client.Bulk.WithRefresh(repo.refresh),
	)
	if err != nil {
		return asset.DiscoveryError{Err: err}
	}
	defer res.Body.Close()
	if res.IsError() {
		return asset.DiscoveryError{Err: fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))}
	}
	return nil
}

func (repo *DiscoveryRepository) DeleteByID(ctx context.Context, ns *namespace.Namespace, assetID string) error {
	if assetID == "" {
		return asset.ErrEmptyID
	}

	return repo.deleteWithQuery(ctx, strings.NewReader(fmt.Sprintf(`{"query":{"term":{"_id": "%s"}}}`, assetID)))
}

func (repo *DiscoveryRepository) DeleteByURN(ctx context.Context, ns *namespace.Namespace, assetURN string) error {
	if assetURN == "" {
		return asset.ErrEmptyURN
	}

	return repo.deleteWithQuery(ctx, strings.NewReader(fmt.Sprintf(`{"query":{"term":{"urn.keyword": "%s"}}}`, assetURN)))
}

func (repo *DiscoveryRepository) deleteWithQuery(ctx context.Context, qry io.Reader) error {
	res, err := repo.cli.client.DeleteByQuery(
		[]string{"_all"},
		qry,
		repo.cli.client.DeleteByQuery.WithContext(ctx),
		repo.cli.client.DeleteByQuery.WithRefresh(true),
	)
	if err != nil {
		return asset.DiscoveryError{Err: fmt.Errorf("error deleting asset: %w", err)}
	}
	defer res.Body.Close()
	if res.IsError() {
		return asset.DiscoveryError{Err: fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))}
	}

	return nil
}

func (repo *DiscoveryRepository) createUpsertBody(ns *namespace.Namespace, ast *asset.Asset) (io.Reader, error) {
	payload := bytes.NewBuffer(nil)
	err := repo.writeInsertAction(payload, ns, ast)
	if err != nil {
		return nil, fmt.Errorf("createBulkInsertPayload: %w", err)
	}

	err = json.NewEncoder(payload).Encode(AssetModel{
		Asset:       *ast,
		NamespaceID: ns.ID.String(),
	})
	fmt.Println(payload.String())
	if err != nil {
		return nil, fmt.Errorf("error serialising asset: %w", err)
	}
	return payload, nil
}

func (repo *DiscoveryRepository) writeInsertAction(w io.Writer, ns *namespace.Namespace, ast *asset.Asset) error {
	action := map[string]interface{}{
		"index": map[string]interface{}{
			"_index": BuildAliasNameFromNamespace(ns),
			"_id":    ast.ID,
		},
	}

	return json.NewEncoder(w).Encode(action)
}
