package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/user"
)

// StarRepository is a type that manages star operation to the primary database
type StarRepository struct {
	client *Client
}

// Create insert a new record in the stars table
func (r *StarRepository) Create(ctx context.Context, userID string, assetID string) (string, error) {
	var starID string
	if userID == "" {
		return "", star.ErrEmptyUserID
	}
	if assetID == "" {
		return "", star.ErrEmptyAssetID
	}

	if err := r.client.db.QueryRowxContext(ctx, `
					INSERT INTO 
					stars 
						(user_id, asset_id)
					VALUES 
						($1, $2)
					RETURNING id
					`, userID, assetID).Scan(&starID); err != nil {
		err := checkPostgresError(err)
		if errors.Is(err, errDuplicateKey) {
			return "", star.DuplicateRecordError{UserID: userID, AssetID: assetID}
		}
		if errors.Is(err, errForeignKeyViolation) {
			return "", star.UserNotFoundError{UserID: userID}
		}
		return "", err
	}
	if starID == "" {
		return "", fmt.Errorf("error star ID is empty from DB")
	}
	return starID, nil
}

// GetStargazers fetch list of user IDs that star an asset
func (r *StarRepository) GetStargazers(ctx context.Context, cfg star.Config, assetID string) ([]user.User, error) {
	if assetID == "" {
		return nil, star.ErrEmptyAssetID
	}

	starCfg := r.buildConfig(cfg)

	var userModels UserModels
	if err := r.client.db.SelectContext(ctx, &userModels, `
		SELECT
			DISTINCT ON (u.id) u.id,
			u.email,
			u.provider,
			u.created_at,
			u.updated_at
		FROM
			stars s
		JOIN
			users u ON s.user_id = u.id
		WHERE
			s.asset_id = $1
		LIMIT
			$2
		OFFSET
			$3
	`, assetID, starCfg.Limit, starCfg.Offset); err != nil {
		return nil, fmt.Errorf("failed fetching users of star: %w", err)
	}

	if len(userModels) == 0 {
		return nil, star.NotFoundError{AssetID: assetID}
	}

	return userModels.toUsers(), nil
}

// GetAllAssetsByUserID fetch list of assets starred by a user
func (r *StarRepository) GetAllAssetsByUserID(ctx context.Context, cfg star.Config, userID string) ([]asset.Asset, error) {
	if userID == "" {
		return nil, star.ErrEmptyUserID
	}

	starCfg := r.buildConfig(cfg)

	var assetModels []AssetModel
	if err := r.client.db.SelectContext(ctx, &assetModels, fmt.Sprintf(`
		SELECT	
			a.id as id,
			a.urn as urn,
			a.type as type,
			a.name as name,
			a.service as service,
			a.description as description,
			a.data as data,
			a.labels as labels,
			a.created_at as created_at,
			a.updated_at as updated_at
		FROM
			stars s
		JOIN
			assets a ON s.asset_id = a.id
		WHERE
			s.user_id = $1
		ORDER BY
			$2 %s
		LIMIT
			$3
		OFFSET
			$4
	`, starCfg.SortDirectionKey), userID, starCfg.SortKey, starCfg.Limit, starCfg.Offset); err != nil {
		return nil, fmt.Errorf("failed fetching stars by user: %w", err)
	}

	if len(assetModels) == 0 {
		return nil, star.NotFoundError{UserID: userID}
	}

	assets := []asset.Asset{}
	for _, am := range assetModels {
		assets = append(assets, am.toAsset())
	}
	return assets, nil
}

// GetAssetByUserID fetch a specific starred asset by user id
func (r *StarRepository) GetAssetByUserID(ctx context.Context, userID string, assetID string) (*asset.Asset, error) {
	if userID == "" {
		return nil, star.ErrEmptyUserID
	}
	if assetID == "" {
		return nil, star.ErrEmptyAssetID
	}

	var asetModel AssetModel
	err := r.client.db.GetContext(ctx, &asetModel, `
		SELECT
			a.id,
			a.urn,
			a.type,
			a.service,
			a.name,
			a.description,
			a.data,
			a.labels,
			a.created_at,
			a.updated_at
		FROM
			stars s
		JOIN
			assets a ON s.asset_id = a.id
		WHERE
			s.user_id = $1 AND s.asset_id = $2
		LIMIT 1
	`, userID, assetID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, star.NotFoundError{AssetID: assetID, UserID: userID}
	}
	if err != nil {
		return nil, fmt.Errorf("failed fetching star by user: %w", err)
	}

	asset := asetModel.toAsset()
	return &asset, nil
}

// Delete will delete/unstar a starred asset for a user id
func (r *StarRepository) Delete(ctx context.Context, userID string, assetID string) error {
	if userID == "" {
		return star.ErrEmptyUserID
	}
	if assetID == "" {
		return star.ErrEmptyAssetID
	}

	res, err := r.client.db.ExecContext(ctx, `
		DELETE FROM
			stars 
		WHERE
			user_id = $1 AND asset_id = $2
	`, userID, assetID)
	if err != nil {
		return fmt.Errorf("failed to unstar an asset: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get row affected  unstarring an asset: %w", err)
	}

	if rowsAffected == 0 {
		return star.NotFoundError{AssetID: assetID, UserID: userID}
	}
	return nil
}

func (r *StarRepository) buildConfig(cfg star.Config) StarConfig {
	sCfg := StarConfig{
		Offset:           0,
		Limit:            DEFAULT_MAX_RESULT_SIZE,
		SortKey:          columnNameCreatedAt,
		SortDirectionKey: sortDirectionDescending,
	}

	if cfg.Size > 0 {
		sCfg.Limit = cfg.Size
	}

	if cfg.Offset < 1 {
		cfg.Offset = 0
	}

	switch cfg.Sort {
	case star.SortKeyCreated:
		sCfg.SortKey = columnNameCreatedAt
	case star.SortKeyUpdated:
		sCfg.SortKey = columnNameUpdatedAt
	}

	switch cfg.SortDirection {
	case star.SortDirectionKeyAscending:
		sCfg.SortDirectionKey = sortDirectionAscending
	case star.SortDirectionKeyDescending:
		sCfg.SortDirectionKey = sortDirectionDescending
	}
	return sCfg
}

// NewStarRepository initializes star repository clients
func NewStarRepository(c *Client) (*StarRepository, error) {
	if c == nil {
		return nil, errNilPostgresClient
	}
	return &StarRepository{
		client: c,
	}, nil
}
