package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/compass/asset"
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
	cli              *elasticsearch.Client
	typeWhiteList    []string
	typeWhiteListSet map[string]bool
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
		return fmt.Errorf("error deleting asset: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}

	return nil
}

func (repo *DiscoveryRepository) GetTypes(ctx context.Context) (map[asset.Type]int, error) {
	resp, err := repo.cli.Cat.Indices(
		repo.cli.Cat.Indices.WithFormat("json"),
		repo.cli.Cat.Indices.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("error from es client %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, fmt.Errorf("error from es server: %s", errorReasonFromResponse(resp))
	}
	var indices []esIndex
	err = json.NewDecoder(resp.Body).Decode(&indices)
	if err != nil {
		return nil, fmt.Errorf("error decoding es response %w", err)
	}

	results := map[asset.Type]int{}
	for _, index := range indices {
		count, err := strconv.Atoi(index.DocsCount)
		if err != nil {
			return results, fmt.Errorf("error converting docs count to a number: %w", err)
		}
		typName := asset.Type(index.Index)
		if !typName.IsValid() {
			continue
		}
		results[typName] = count
	}

	return results, nil
}

func (repo *DiscoveryRepository) createUpsertBody(ast asset.Asset) (io.Reader, error) {
	payload := bytes.NewBuffer(nil)
	err := repo.writeInsertAction(payload, ast)
	if err != nil {
		return nil, fmt.Errorf("createBulkInsertPayload: %w", err)
	}

	err = json.NewEncoder(payload).Encode(ast)
	if err != nil {
		return nil, fmt.Errorf("error serialising asset: %w", err)
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
