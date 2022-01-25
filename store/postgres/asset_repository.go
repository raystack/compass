package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/odpf/columbus/asset"
)

const (
	DEFAULT_MAX_SIZE = 100
)

// AssetRepository is a type that manages user operation to the primary database
type AssetRepository struct {
	client            *Client
	defaultGetMaxSize int
}

// Get retrieves list of assets with filters via config
func (r *AssetRepository) Get(ctx context.Context, config asset.GetConfig) (assets []asset.Asset, err error) {
	query, args := r.buildGetQuery(config)
	ams := []*Asset{}
	err = r.client.db.SelectContext(ctx, &ams, query, args...)
	if err != nil {
		err = fmt.Errorf("error getting asset list: %w", err)
		return
	}

	assets = []asset.Asset{}
	for _, am := range ams {
		assets = append(assets, am.toDomain())
	}

	return
}

// GetByID retrieves asset by its ID
func (r *AssetRepository) GetByID(ctx context.Context, id string) (asset.Asset, error) {
	query := `SELECT * FROM assets WHERE id = $1 LIMIT 1;`

	am := &Asset{}
	err := r.client.db.GetContext(ctx, am, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{AssetID: id}
	}
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error getting asset with ID = \"%s\" from DB: %w", id, err)
	}

	return am.toDomain(), nil
}

// Upsert creates a new asset if it does not exist yet.
// It updates if asset does exist.
// Checking existance is done using "urn", "type", and "service" fields.
func (r *AssetRepository) Upsert(ctx context.Context, ast *asset.Asset) error {
	assetID, err := r.getID(ctx, ast)
	if err != nil {
		return fmt.Errorf("error getting asset ID: %w", err)
	}
	if assetID == "" {
		assetID, err = r.insert(ctx, ast)
		if err != nil {
			return fmt.Errorf("error inserting asset to DB: %w", err)
		}
	} else {
		err = r.update(ctx, assetID, ast)
		if err != nil {
			return fmt.Errorf("error updating asset to DB: %w", err)
		}
	}

	ast.ID = assetID
	return nil
}

// Delete removes asset using its ID
func (r *AssetRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM assets WHERE id = $1;`
	res, err := r.client.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting asset with ID = \"%s\": %w", id, err)
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %w", err)
	}
	if affectedRows == 0 {
		return asset.NotFoundError{AssetID: id}
	}

	return nil
}

func (r *AssetRepository) buildGetQuery(config asset.GetConfig) (query string, args []interface{}) {
	whereFields := []string{}
	args = []interface{}{}

	if config.Type != "" {
		whereFields = append(whereFields, "type")
		args = append(args, config.Type)
	}
	if config.Service != "" {
		whereFields = append(whereFields, "service")
		args = append(args, config.Service)
	}
	size := config.Size
	if size == 0 {
		size = r.defaultGetMaxSize
	}

	args = append(args, size, config.Offset)

	whereClauses := []string{}
	for i, field := range whereFields {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", field, i+1))
	}
	totalWhereClauses := len(whereClauses)

	query = "SELECT * FROM assets "
	if totalWhereClauses > 0 {
		query += "WHERE " + strings.Join(whereClauses, " AND ") + " "
	}
	query += fmt.Sprintf("LIMIT $%d OFFSET $%d;", totalWhereClauses+1, totalWhereClauses+2)

	return
}

func (r *AssetRepository) insert(ctx context.Context, ast *asset.Asset) (id string, err error) {
	err = r.client.db.QueryRowxContext(ctx,
		`INSERT INTO assets 
			(urn, type, service, name, description, data, labels, created_at, updated_at)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		ast.URN, ast.Type, ast.Service, ast.Name, ast.Description, ast.Data, ast.Labels, ast.CreatedAt, ast.UpdatedAt).Scan(&id)

	return
}

func (r *AssetRepository) update(ctx context.Context, id string, ast *asset.Asset) error {
	_, err := r.client.db.ExecContext(ctx,
		`UPDATE assets
		SET urn = $1,
			type = $2,
			service = $3,
			name = $4,
			description = $5,
			data = $6,
			labels = $7,
			updated_at = $8
		WHERE id = $9;
		`,
		ast.URN, ast.Type, ast.Service, ast.Name, ast.Description, ast.Data, ast.Labels, ast.UpdatedAt, id)
	return err
}

func (r *AssetRepository) getID(ctx context.Context, ast *asset.Asset) (id string, err error) {
	query := `SELECT id FROM assets WHERE urn = $1 AND type = $2 AND service = $3;`
	err = r.client.db.GetContext(ctx, &id, query, ast.URN, ast.Type, ast.Service)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		err = fmt.Errorf(
			"error getting asset's ID with urn = \"%s\", type = \"%s\", service = \"%s\": %w",
			ast.URN, ast.Type, ast.Service, err)
	}

	return
}

// NewAssetRepository initializes user repository clients
func NewAssetRepository(c *Client, defaultGetMaxSize int) (*AssetRepository, error) {
	if c == nil {
		return nil, errors.New("postgres client is nil")
	}
	if defaultGetMaxSize == 0 {
		defaultGetMaxSize = DEFAULT_MAX_SIZE
	}

	return &AssetRepository{
		client:            c,
		defaultGetMaxSize: defaultGetMaxSize,
	}, nil
}
