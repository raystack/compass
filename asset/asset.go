package asset

import (
	"context"
	"time"

	"github.com/odpf/columbus/user"
)

// Asset is a model that wraps arbitrary data with Columbus' context
type Asset struct {
	ID          string                 `json:"id"`
	URN         string                 `json:"urn"`
	Type        Type                   `json:"type"`
	Name        string                 `json:"name"`
	Service     string                 `json:"service"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Labels      map[string]string      `json:"labels"`
	Owners      []user.User            `json:"owners"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type Repository interface {
	Get(context.Context, GetConfig) ([]Asset, error)
	GetByID(ctx context.Context, id string) (Asset, error)
	Upsert(context.Context, *Asset) error
	Delete(ctx context.Context, id string) error
}

type GetConfig struct {
	Text    string `json:"text"`
	Type    Type   `json:"type"`
	Service string `json:"service"`
	Size    int    `json:"size"`
	Offset  int    `json:"offset"`
}
