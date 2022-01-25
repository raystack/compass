package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/record"
)

// TypeRepository is an implementation of record.TypeRepository
// that uses elasticsearch as a backing store
type TypeRepository struct {
	cli *elasticsearch.Client
}

func (repo *TypeRepository) GetAll(ctx context.Context) (map[record.TypeName]int, error) {
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

	results := map[record.TypeName]int{}
	for _, index := range indices {
		count, err := strconv.Atoi(index.DocsCount)
		if err != nil {
			return results, fmt.Errorf("error converting docs count to a number: %w", err)
		}
		typName := record.TypeName(index.Index)
		if err := typName.IsValid(); err != nil {
			continue
		}
		results[typName] = count
	}

	return results, nil
}

func NewTypeRepository(cli *elasticsearch.Client) *TypeRepository {
	return &TypeRepository{
		cli: cli,
	}
}
