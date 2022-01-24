package asset

import (
	"context"
	"time"
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
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type Repository interface {
	Get(context.Context, GetConfig) ([]Asset, error)
	GetByID(ctx context.Context, id string) (Asset, error)
	Create(context.Context, *Asset) error
	Update(context.Context, *Asset) error
	Delete(ctx context.Context, id string) error
}

type GetConfig struct {
	Text   string `json:"text"`
	Size   int    `json:"size"`
	Offset int    `json:"offset"`
}
