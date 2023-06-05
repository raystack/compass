package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/star"
	"github.com/goto/compass/core/user"
)

type StarClauses struct {
	Limit            int
	Offset           int
	SortKey          string
	SortDirectionKey string
}

// StarRepository is a type that manages star operation to the primary database
type StarRepository struct {
	client *Client
}

// Create insert a new record in the stars table
func (r *StarRepository) Create(ctx context.Context, userID, assetID string) (string, error) {
	var starID string
	if userID == "" {
		return "", star.ErrEmptyUserID
	}
	if assetID == "" {
		return "", star.ErrEmptyAssetID
	}

	if !isValidUUID(userID) {
		return "", star.InvalidError{UserID: userID}
	}

	if !isValidUUID(assetID) {
		return "", star.InvalidError{AssetID: assetID}
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
func (r *StarRepository) GetStargazers(ctx context.Context, flt star.Filter, assetID string) ([]user.User, error) {
	if assetID == "" {
		return nil, star.ErrEmptyAssetID
	}

	if !isValidUUID(assetID) {
		return nil, star.InvalidError{AssetID: assetID}
	}

	starClausesValue := r.buildClausesValue(flt)
	var userModels UserModels
	if err := r.client.db.SelectContext(ctx, &userModels, `
		SELECT
			DISTINCT ON (u.id) u.id,
      u.uuid,
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
		LIMIT $2
		OFFSET $3
	`, assetID, starClausesValue.Limit, starClausesValue.Offset); err != nil {
		return nil, fmt.Errorf("failed fetching users of star: %w", err)
	}

	if len(userModels) == 0 {
		return nil, star.NotFoundError{AssetID: assetID}
	}

	return userModels.toUsers(), nil
}

// GetAllAssetsByUserID fetch list of assets starred by a user
func (r *StarRepository) GetAllAssetsByUserID(ctx context.Context, flt star.Filter, userID string) ([]asset.Asset, error) {
	if userID == "" {
		return nil, star.ErrEmptyUserID
	}

	if !isValidUUID(userID) {
		return nil, star.InvalidError{UserID: userID}
	}

	starClausesValue := r.buildClausesValue(flt)

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
			a.version as version,
			a.created_at as created_at,
			a.updated_at as updated_at,
			u.id as "updated_by.id",
			u.uuid as "updated_by.uuid",
			u.email as "updated_by.email",
			u.provider as "updated_by.provider",
			u.created_at as "updated_by.created_at",
			u.updated_at as "updated_by.updated_at"
		FROM
			stars s
		INNER JOIN
			assets a ON s.asset_id = a.id
		LEFT JOIN
			users u ON a.updated_by = u.id
		WHERE
			s.user_id = $1
		ORDER BY
			$2 %s
		LIMIT
			$3
		OFFSET
			$4
	`, starClausesValue.SortDirectionKey), userID, starClausesValue.SortKey, starClausesValue.Limit, starClausesValue.Offset); err != nil {
		return nil, fmt.Errorf("failed fetching stars by user: %w", err)
	}

	if len(assetModels) == 0 {
		return nil, star.NotFoundError{UserID: userID}
	}

	assets := []asset.Asset{}
	for _, am := range assetModels {
		assets = append(assets, am.toAsset(nil))
	}
	return assets, nil
}

// GetAssetByUserID fetch a specific starred asset by user id
func (r *StarRepository) GetAssetByUserID(ctx context.Context, userID, assetID string) (asset.Asset, error) {
	if userID == "" {
		return asset.Asset{}, star.ErrEmptyUserID
	}
	if assetID == "" {
		return asset.Asset{}, star.ErrEmptyAssetID
	}

	if !isValidUUID(userID) {
		return asset.Asset{}, star.InvalidError{UserID: userID}
	}
	if !isValidUUID(assetID) {
		return asset.Asset{}, star.InvalidError{AssetID: assetID}
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
			a.version,
			a.created_at,
			a.updated_at,
			u.id as "updated_by.id",
			u.uuid as "updated_by.uuid",
			u.email as "updated_by.email",
			u.provider as "updated_by.provider",
			u.created_at as "updated_by.created_at",
			u.updated_at as "updated_by.updated_at"
		FROM
			stars s
		INNER JOIN
			assets a ON s.asset_id = a.id
		LEFT JOIN
			users u ON a.updated_by = u.id
		WHERE
			s.user_id = $1 AND s.asset_id = $2
		LIMIT 1
	`, userID, assetID)
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, star.NotFoundError{AssetID: assetID, UserID: userID}
	}
	if err != nil {
		return asset.Asset{}, fmt.Errorf("failed fetching star by user: %w", err)
	}

	asset := asetModel.toAsset(nil)
	return asset, nil
}

// Delete will delete/unstar a starred asset for a user id
func (r *StarRepository) Delete(ctx context.Context, userID, assetID string) error {
	if userID == "" {
		return star.ErrEmptyUserID
	}
	if assetID == "" {
		return star.ErrEmptyAssetID
	}

	if !isValidUUID(userID) {
		return star.InvalidError{UserID: userID}
	}
	if !isValidUUID(assetID) {
		return star.InvalidError{AssetID: assetID}
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
		return fmt.Errorf("failed to get row affected unstarring an asset: %w", err)
	}

	if rowsAffected == 0 {
		return star.NotFoundError{AssetID: assetID, UserID: userID}
	}
	return nil
}

func (r *StarRepository) buildClausesValue(flt star.Filter) StarClauses {
	sCfg := StarClauses{
		Offset:           0,
		Limit:            DEFAULT_MAX_RESULT_SIZE,
		SortKey:          columnNameCreatedAt,
		SortDirectionKey: sortDirectionDescending,
	}

	if flt.Size > 0 {
		sCfg.Limit = flt.Size
	}

	if flt.Offset < 1 {
		flt.Offset = 0
	}

	switch flt.Sort {
	case star.SortKeyCreated:
		sCfg.SortKey = columnNameCreatedAt
	case star.SortKeyUpdated:
		sCfg.SortKey = columnNameUpdatedAt
	}

	switch flt.SortDirection {
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
