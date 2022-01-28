package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/asset"
)

const (
	defaultGetSize   = 20
	defaultSortField = "name.keyword"

	defaultScrollTimeout   = 30 * time.Second
	defaultScrollBatchSize = 20
)

var indexTypeMap = map[asset.Type]string{
	asset.TypeTable:     "table",
	asset.TypeTopic:     "topic",
	asset.TypeJob:       "job",
	asset.TypeDashboard: "dashboard",
}

// DiscoveryRepository implements discovery.Repository
// with elasticsearch as the backing store.
type DiscoveryRepository struct {
	cli *elasticsearch.Client
}

func NewDiscoveryRepository(cli *elasticsearch.Client) *DiscoveryRepository {
	return &DiscoveryRepository{
		cli: cli,
	}
}

func (repo *DiscoveryRepository) Upsert(ctx context.Context, ast asset.Asset) error {
	if ast.ID == "" {
		return asset.ErrEmptyID
	}
	if !ast.Type.IsValid() {
		return asset.ErrUnknownType
	}

	body, err := repo.createUpsertBody(ast)
	if err != nil {
		return fmt.Errorf("error serialising payload: %w", err)
	}
	res, err := repo.cli.Bulk(
		body,
		repo.cli.Bulk.WithRefresh("true"),
		repo.cli.Bulk.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}
	return nil
}

func (repo *DiscoveryRepository) Delete(ctx context.Context, assetID string) error {
	if assetID == "" {
		return asset.ErrEmptyID
	}

	res, err := repo.cli.DeleteByQuery(
		[]string{"_all"},
		strings.NewReader(fmt.Sprintf(`{"query":{"terms":{"_id": ["%s"]}}}`, assetID)),
		repo.cli.DeleteByQuery.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error deleting record: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}

	return nil
}

func (repo *DiscoveryRepository) createUpsertBody(ast asset.Asset) (io.Reader, error) {
	payload := bytes.NewBuffer(nil)
	err := repo.writeInsertAction(payload, ast)
	if err != nil {
		return nil, fmt.Errorf("createBulkInsertPayload: %w", err)
	}
	err = json.NewEncoder(payload).Encode(ast)
	if err != nil {
		return nil, fmt.Errorf("error serialising record: %w", err)
	}
	return payload, nil
}

func (repo *DiscoveryRepository) writeInsertAction(w io.Writer, ast asset.Asset) error {
	action := map[string]interface{}{
		"index": map[string]interface{}{
			"_index": indexTypeMap[ast.Type],
			"_id":    ast.ID,
		},
	}

	return json.NewEncoder(w).Encode(action)
}
