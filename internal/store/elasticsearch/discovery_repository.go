package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/goto/compass/core/asset"
	"github.com/goto/salt/log"
)

// DiscoveryRepository implements discovery.Repository
// with elasticsearch as the backing store.
type DiscoveryRepository struct {
	cli    *Client
	logger log.Logger
}

func NewDiscoveryRepository(cli *Client, logger log.Logger) *DiscoveryRepository {
	return &DiscoveryRepository{
		cli:    cli,
		logger: logger,
	}
}

func (repo *DiscoveryRepository) Upsert(ctx context.Context, ast asset.Asset) error {
	if ast.ID == "" {
		return asset.ErrEmptyID
	}
	if !ast.Type.IsValid() {
		return asset.ErrUnknownType
	}

	idxExists, err := repo.cli.indexExists(ctx, "Upsert", ast.Service)
	if err != nil {
		return asset.DiscoveryError{
			Op:    "IndexExists",
			ID:    ast.ID,
			Index: ast.Service,
			Err:   err,
		}
	}

	if !idxExists {
		if err := repo.cli.CreateIdx(ctx, "Upsert", ast.Service); err != nil {
			return asset.DiscoveryError{
				Op:    "CreateIndex",
				ID:    ast.ID,
				Index: ast.Service,
				Err:   err,
			}
		}
	}

	return repo.indexAsset(ctx, ast)
}

func (repo *DiscoveryRepository) DeleteByID(ctx context.Context, assetID string) error {
	if assetID == "" {
		return asset.ErrEmptyID
	}

	return repo.deleteWithQuery(ctx, "DeleteByID", fmt.Sprintf(`{"query":{"term":{"_id": %q}}}`, assetID))
}

func (repo *DiscoveryRepository) DeleteByURN(ctx context.Context, assetURN string) error {
	if assetURN == "" {
		return asset.ErrEmptyURN
	}

	return repo.deleteWithQuery(ctx, "DeleteByURN", fmt.Sprintf(`{"query":{"term":{"urn.keyword": %q}}}`, assetURN))
}

func (repo *DiscoveryRepository) deleteWithQuery(ctx context.Context, discoveryOp, qry string) (err error) {
	defer func(start time.Time) {
		const op = "delete_by_query"
		repo.cli.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: discoveryOp,
			start:       start,
			err:         err,
		})
	}(time.Now())

	deleteByQ := repo.cli.client.DeleteByQuery
	res, err := deleteByQ(
		[]string{defaultSearchIndex},
		strings.NewReader(qry),
		deleteByQ.WithContext(ctx),
		deleteByQ.WithRefresh(true),
		deleteByQ.WithIgnoreUnavailable(true),
	)
	if err != nil {
		return asset.DiscoveryError{
			Op:  "DeleteDoc",
			Err: fmt.Errorf("query: %s: %w", qry, err),
		}
	}

	defer drainBody(res)
	if res.IsError() {
		code, reason := errorCodeAndReason(res)
		return asset.DiscoveryError{
			Op:     "DeleteDoc",
			ESCode: code,
			Err:    fmt.Errorf("query: %s: %s", qry, reason),
		}
	}

	return nil
}

func (repo *DiscoveryRepository) indexAsset(ctx context.Context, ast asset.Asset) (err error) {
	defer func(start time.Time) {
		const op = "index"
		repo.cli.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: "Upsert",
			start:       start,
			err:         err,
		})
	}(time.Now())

	body, err := createUpsertBody(ast)
	if err != nil {
		return asset.DiscoveryError{
			Op:  "EncodeAsset",
			ID:  ast.ID,
			Err: err,
		}
	}

	index := repo.cli.client.Index
	resp, err := index(
		ast.Service,
		body,
		index.WithDocumentID(url.PathEscape(ast.ID)),
		index.WithContext(ctx),
	)
	if err != nil {
		return asset.DiscoveryError{
			Op:    "IndexDoc",
			ID:    ast.ID,
			Index: ast.Service,
			Err:   err,
		}
	}
	defer drainBody(resp)

	if resp.IsError() {
		code, reason := errorCodeAndReason(resp)
		return asset.DiscoveryError{
			Op:     "IndexDoc",
			ID:     ast.ID,
			Index:  ast.Service,
			ESCode: code,
			Err:    errors.New(reason),
		}
	}

	return nil
}

func createUpsertBody(ast asset.Asset) (io.Reader, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(ast); err != nil {
		return nil, fmt.Errorf("encode asset: %w", err)
	}

	return &buf, nil
}
