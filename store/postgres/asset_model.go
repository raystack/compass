package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/odpf/columbus/asset"
)

type Asset struct {
	ID          string    `db:"id"`
	URN         string    `db:"urn"`
	Type        string    `db:"type"`
	Name        string    `db:"name"`
	Service     string    `db:"service"`
	Description string    `db:"description"`
	Data        JSONMap   `db:"data"`
	Labels      JSONMap   `db:"labels"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (a *Asset) toAsset() asset.Asset {
	return asset.Asset{
		ID:          a.ID,
		URN:         a.URN,
		Type:        asset.Type(a.Type),
		Name:        a.Name,
		Service:     a.Service,
		Description: a.Description,
		Data:        a.Data,
		Labels:      a.buildLabels(),
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func (a *Asset) buildLabels() map[string]string {
	if a.Labels == nil {
		return nil
	}

	result := make(map[string]string)
	for key, value := range a.Labels {
		strKey := fmt.Sprintf("%v", key)
		strValue := fmt.Sprintf("%v", value)

		result[strKey] = strValue
	}

	return result
}

func newAssetModel(a asset.Asset) *Asset {
	labels := make(map[string]interface{})
	for key, value := range a.Labels {
		labels[key] = value
	}

	return &Asset{
		ID:          a.ID,
		URN:         a.URN,
		Type:        a.Type.String(),
		Name:        a.Name,
		Service:     a.Service,
		Description: a.Description,
		Data:        a.Data,
		Labels:      labels,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	ba, err := m.MarshalJSON()
	return string(ba), err
}

func (m *JSONMap) Scan(value interface{}) error {
	var ba []byte
	switch v := value.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	t := map[string]interface{}{}
	err := json.Unmarshal(ba, &t)
	*m = JSONMap(t)
	return err
}

// MarshalJSON to output non base64 encoded []byte
func (m JSONMap) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	t := (map[string]interface{})(m)
	return json.Marshal(t)
}

// UnmarshalJSON to deserialize []byte
func (m *JSONMap) UnmarshalJSON(b []byte) error {
	t := map[string]interface{}{}
	err := json.Unmarshal(b, &t)
	*m = JSONMap(t)
	return err
}
