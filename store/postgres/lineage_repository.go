package postgres

import (
	"context"
	"errors"

	"github.com/odpf/columbus/lineage"
)

type LineageRepository struct {
	client *Client
}

// NewLineageRepository initializes tag repository
// all methods in tag repository uses passed by reference
// which will mutate the reference variable in method's argument
func NewLineageRepository(client *Client) (*LineageRepository, error) {
	if client == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &LineageRepository{
		client: client,
	}, nil
}

func (repo *LineageRepository) GetEdges(ctx context.Context) ([]lineage.Edge, error) {
	return nil, nil
}
