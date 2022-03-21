package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/user"
	"github.com/r3labs/diff/v2"
)

// AssetRepository is a type that manages user operation to the primary database
type AssetRepository struct {
	client              *Client
	userRepo            *UserRepository
	defaultGetMaxSize   int
	defaultUserProvider string
}

// GetAll retrieves list of assets with filters via config
func (r *AssetRepository) GetAll(ctx context.Context, config asset.Config) ([]asset.Asset, error) {
	size := config.Size
	if size == 0 {
		size = r.defaultGetMaxSize
	}

	builder := r.getAssetSQL().
		Limit(uint64(size)).
		Offset(uint64(config.Offset))
	builder = r.buildFilterQuery(builder, config)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}
	ams := []*AssetModel{}
	err = r.client.db.SelectContext(ctx, &ams, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting asset list: %w", err)
	}

	assets := []asset.Asset{}
	for _, am := range ams {
		assets = append(assets, am.toAsset(nil))
	}

	return assets, nil
}

// GetCount retrieves number of assets for every type
func (r *AssetRepository) GetCount(ctx context.Context, config asset.Config) (total int, err error) {
	builder := sq.Select("count(1)").From("assets")
	builder = r.buildFilterQuery(builder, config)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		err = fmt.Errorf("error building count query: %w", err)
		return
	}
	err = r.client.db.GetContext(ctx, &total, query, args...)
	if err != nil {
		err = fmt.Errorf("error getting asset list: %w", err)
	}

	return
}

// GetByID retrieves asset by its ID
func (r *AssetRepository) GetByID(ctx context.Context, id string) (ast asset.Asset, err error) {
	if !isValidUUID(id) {
		err = asset.InvalidError{AssetID: id}
		return
	}

	builder := r.getAssetSQL().
		Where(sq.Eq{"a.id": id}).
		Limit(1)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		err = fmt.Errorf("error building query: %w", err)
		return
	}

	am := &AssetModel{}
	err = r.client.db.GetContext(ctx, am, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		err = asset.NotFoundError{AssetID: id}
		return
	}

	if err != nil {
		err = fmt.Errorf("error getting asset with ID = \"%s\": %w", id, err)
		return
	}

	owners, err := r.getOwners(ctx, id)
	if err != nil {
		err = fmt.Errorf("error getting asset with ID = \"%s\": %w", id, err)
		return
	}

	ast = am.toAsset(owners)

	return
}

func (r *AssetRepository) Find(ctx context.Context, assetURN string, assetType asset.Type, assetService string) (ast asset.Asset, err error) {
	builder := r.getAssetSQL().
		Where(sq.Eq{
			"a.urn":     assetURN,
			"a.type":    assetType,
			"a.service": assetService,
		})
	query, args, err := r.buildSQL(builder)
	if err != nil {
		err = fmt.Errorf("error building query: %w", err)
		return
	}

	var assetModel AssetModel
	if err = r.client.db.GetContext(ctx, &assetModel, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = asset.NotFoundError{}
			return
		}
		err = fmt.Errorf(
			"error getting asset with urn = \"%s\", type = \"%s\", service = \"%s\": %w",
			assetURN, assetType, assetService, err)
		return
	}

	owners, err := r.getOwners(ctx, assetModel.ID)
	if err != nil {
		err = fmt.Errorf("error getting asset's current owners: %w", err)
		return
	}

	ast = assetModel.toAsset(owners)

	return
}

// GetVersionHistory retrieves the previous versions of an asset
func (r *AssetRepository) GetVersionHistory(ctx context.Context, cfg asset.Config, id string) (avs []asset.AssetVersion, err error) {
	if !isValidUUID(id) {
		err = asset.InvalidError{AssetID: id}
		return
	}

	size := cfg.Size
	if size == 0 {
		size = r.defaultGetMaxSize
	}

	builder := r.getAssetVersionSQL().
		Where(sq.Eq{"a.asset_id": id}).
		OrderBy("version DESC").
		Limit(uint64(size)).
		Offset(uint64(cfg.Offset))
	query, args, err := r.buildSQL(builder)
	if err != nil {
		err = fmt.Errorf("error building query: %w", err)
		return
	}

	var assetModels []AssetModel
	err = r.client.db.SelectContext(ctx, &assetModels, query, args...)
	if err != nil {
		err = fmt.Errorf("failed fetching last versions: %w", err)
		return
	}

	if len(assetModels) == 0 {
		err = asset.NotFoundError{AssetID: id}
		return
	}

	for _, am := range assetModels {
		av, ferr := am.toAssetVersion()
		if ferr != nil {
			err = fmt.Errorf("failed converting asset model to asset version: %w", ferr)
			return
		}
		avs = append(avs, av)
	}

	return avs, nil
}

// GetByVersion retrieves the specific asset version
func (r *AssetRepository) GetByVersion(ctx context.Context, id string, version string) (ast asset.Asset, err error) {
	if !isValidUUID(id) {
		err = asset.InvalidError{AssetID: id}
		return
	}

	latestAsset, err := r.GetByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		err = asset.NotFoundError{AssetID: id}
		return
	}

	if err != nil {
		return
	}

	if latestAsset.Version == version {
		ast = latestAsset
		return
	}

	var assetModel AssetModel
	builder := r.getAssetVersionSQL().
		Where(sq.Eq{"a.asset_id": id, "a.version": version})
	query, args, err := r.buildSQL(builder)
	if err != nil {
		err = fmt.Errorf("error building query: %w", err)
		return
	}

	err = r.client.db.GetContext(ctx, &assetModel, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		err = asset.NotFoundError{AssetID: id}
		return
	}

	if err != nil {
		err = fmt.Errorf("failed fetching asset version: %w", err)
		return
	}

	ast, err = assetModel.toVersionedAsset(latestAsset)

	return
}

// Upsert creates a new asset if it does not exist yet.
// It updates if asset does exist.
// Checking existance is done using "urn", "type", and "service" fields.
func (r *AssetRepository) Upsert(ctx context.Context, ast *asset.Asset) (string, error) {
	fetchedAsset, err := r.Find(ctx, ast.URN, ast.Type, ast.Service)
	if errors.As(err, new(asset.NotFoundError)) {
		err = nil
	}
	if err != nil {
		return "", fmt.Errorf("error getting asset by URN: %w", err)
	}

	if fetchedAsset.ID == "" {
		// insert flow
		fetchedAsset.ID, err = r.insert(ctx, ast)
		if err != nil {
			return fetchedAsset.ID, fmt.Errorf("error inserting asset to DB: %w", err)
		}
	} else {
		// update flow
		changelog, err := fetchedAsset.Diff(ast)
		if err != nil {
			return "", fmt.Errorf("error diffing two assets: %w", err)
		}

		err = r.update(ctx, fetchedAsset.ID, ast, &fetchedAsset, changelog)
		if err != nil {
			return "", fmt.Errorf("error updating asset to DB: %w", err)
		}
	}

	return fetchedAsset.ID, nil
}

// Delete removes asset using its ID
func (r *AssetRepository) Delete(ctx context.Context, id string) error {

	if !isValidUUID(id) {
		return asset.InvalidError{AssetID: id}
	}

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

func (r *AssetRepository) buildFilterQuery(builder sq.SelectBuilder, config asset.Config) sq.SelectBuilder {
	clause := sq.Eq{}
	if config.Type != "" {
		clause["type"] = config.Type
	}
	if config.Service != "" {
		clause["service"] = config.Service
	}

	if len(clause) > 0 {
		builder = builder.Where(clause)
	}
	return builder
}

func (r *AssetRepository) insert(ctx context.Context, ast *asset.Asset) (id string, err error) {
	err = r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		query, args, err := sq.Insert("assets").
			Columns("urn", "type", "service", "name", "description", "data", "labels", "updated_by", "version").
			Values(ast.URN, ast.Type, ast.Service, ast.Name, ast.Description, ast.Data, ast.Labels, ast.UpdatedBy.ID, asset.BaseVersion).
			Suffix("RETURNING \"id\"").
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return fmt.Errorf("error building insert query: %w", err)
		}

		err = tx.QueryRowContext(ctx, query, args...).Scan(&id)
		if err != nil {
			return fmt.Errorf("error running insert query: %w", err)
		}

		users, err := r.createOrFetchUserIDs(ctx, tx, ast.Owners)
		if err != nil {
			return fmt.Errorf("error creating and fetching owners: %w", err)
		}

		err = r.insertOwners(ctx, tx, id, users)
		if err != nil {
			return fmt.Errorf("error running insert owners query: %w", err)
		}

		return nil
	})

	return
}

func (r *AssetRepository) update(ctx context.Context, assetID string, newAsset *asset.Asset, oldAsset *asset.Asset, clog diff.Changelog) error {

	if !isValidUUID(assetID) {
		return asset.InvalidError{AssetID: assetID}
	}

	if len(clog) == 0 {
		return nil
	}

	return r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		// update assets
		newVersion, err := asset.IncreaseMinorVersion(oldAsset.Version)
		if err != nil {
			return err
		}

		err = r.execContext(ctx, tx,
			`UPDATE assets
			SET urn = $1,
				type = $2,
				service = $3,
				name = $4,
				description = $5,
				data = $6,
				labels = $7,
				updated_at = $8,
				updated_by = $9,
				version = $10
			WHERE id = $11;
			`,
			newAsset.URN, newAsset.Type, newAsset.Service, newAsset.Name, newAsset.Description, newAsset.Data, newAsset.Labels, time.Now(), newAsset.UpdatedBy.ID, newVersion, assetID)
		if err != nil {
			return fmt.Errorf("error running update asset query: %w", err)
		}

		// insert versions
		if err = r.insertAssetVersion(ctx, tx, oldAsset, clog); err != nil {
			return err
		}

		// managing owners
		newAssetOwners, err := r.createOrFetchUserIDs(ctx, tx, newAsset.Owners)
		if err != nil {
			return fmt.Errorf("error creating and fetching owners: %w", err)
		}
		toInserts, toRemoves := r.compareOwners(oldAsset.Owners, newAssetOwners)
		if err := r.insertOwners(ctx, tx, assetID, toInserts); err != nil {
			return fmt.Errorf("error inserting asset's new owners: %w", err)
		}
		if err := r.removeOwners(ctx, tx, assetID, toRemoves); err != nil {
			return fmt.Errorf("error removing asset's old owners: %w", err)
		}

		return nil
	})
}

func (r *AssetRepository) insertAssetVersion(ctx context.Context, execer sqlx.ExecerContext, oldAsset *asset.Asset, clog diff.Changelog) (err error) {
	if oldAsset == nil {
		err = asset.ErrNilAsset
		return
	}

	if clog == nil {
		err = fmt.Errorf("changelog is nil when insert to asset version")
		return
	}

	jsonChangelog, err := json.Marshal(clog)
	if err != nil {
		return err
	}
	query, args, err := sq.Insert("assets_versions").
		Columns("asset_id", "urn", "type", "service", "name", "description", "data", "labels", "created_at", "updated_at", "updated_by", "version", "owners", "changelog").
		Values(oldAsset.ID, oldAsset.URN, oldAsset.Type, oldAsset.Service, oldAsset.Name, oldAsset.Description, oldAsset.Data, oldAsset.Labels,
			oldAsset.CreatedAt, oldAsset.UpdatedAt, oldAsset.UpdatedBy.ID, oldAsset.Version, oldAsset.Owners, jsonChangelog).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building insert query: %w", err)
	}

	if err = r.execContext(ctx, execer, query, args...); err != nil {
		return fmt.Errorf("error running insert asset version query: %w", err)
	}

	return
}

func (r *AssetRepository) getOwners(ctx context.Context, assetID string) (owners []user.User, err error) {

	if !isValidUUID(assetID) {
		return nil, asset.InvalidError{AssetID: assetID}
	}

	query := `
		SELECT
			u.id as "id",
			u.email as "email",
			u.provider as "provider"
		FROM asset_owners ao
		JOIN users u on ao.user_id = u.id
		WHERE asset_id = $1`

	err = r.client.db.SelectContext(ctx, &owners, query, assetID)
	if err != nil {
		err = fmt.Errorf("error getting asset's owners: %w", err)
	}

	return
}

func (r *AssetRepository) insertOwners(ctx context.Context, execer sqlx.ExecerContext, assetID string, owners []user.User) (err error) {
	if len(owners) == 0 {
		return
	}

	if !isValidUUID(assetID) {
		return asset.InvalidError{AssetID: assetID}
	}

	var values []string
	var args = []interface{}{assetID}
	for i, owner := range owners {
		values = append(values, fmt.Sprintf("($1, $%d)", i+2))
		args = append(args, owner.ID)
	}
	query := fmt.Sprintf(`
		INSERT INTO asset_owners
			(asset_id, user_id)
		VALUES %s`, strings.Join(values, ","))
	err = r.execContext(ctx, execer, query, args...)
	if err != nil {
		err = fmt.Errorf("error running insert owners query: %w", err)
	}

	return
}

func (r *AssetRepository) removeOwners(ctx context.Context, execer sqlx.ExecerContext, assetID string, owners []user.User) (err error) {
	if len(owners) == 0 {
		return
	}

	if !isValidUUID(assetID) {
		return asset.InvalidError{AssetID: assetID}
	}

	var user_ids []string
	var args = []interface{}{assetID}
	for i, owner := range owners {
		user_ids = append(user_ids, fmt.Sprintf("$%d", i+2))
		args = append(args, owner.ID)
	}
	query := fmt.Sprintf(
		`DELETE FROM asset_owners WHERE asset_id = $1 AND user_id in (%s)`,
		strings.Join(user_ids, ","),
	)
	err = r.execContext(ctx, execer, query, args...)
	if err != nil {
		err = fmt.Errorf("error running delete owners query: %w", err)
	}

	return
}

func (r *AssetRepository) compareOwners(current, newOwners []user.User) (toInserts, toRemove []user.User) {
	if len(current) == 0 && len(newOwners) == 0 {
		return
	}

	currMap := map[string]int{}
	for _, curr := range current {
		currMap[curr.ID] = 1
	}

	for _, n := range newOwners {
		_, exists := currMap[n.ID]
		if exists {
			// if exists, it means that both new and current have it.
			// we remove it from the map,
			// so that what's left in the map is the that only exists in current
			// and have to be removed
			delete(currMap, n.ID)
		} else {
			toInserts = append(toInserts, user.User{ID: n.ID})
		}
	}

	for id := range currMap {
		toRemove = append(toRemove, user.User{ID: id})
	}

	return
}

func (r *AssetRepository) createOrFetchUserIDs(ctx context.Context, tx *sqlx.Tx, users []user.User) (results []user.User, err error) {
	for _, u := range users {
		if u.ID != "" {
			results = append(results, u)
			continue
		}
		var userID string
		userID, err = r.userRepo.GetID(ctx, u.Email)
		if errors.As(err, &user.NotFoundError{}) {
			u.Provider = r.defaultUserProvider
			userID, err = r.userRepo.CreateWithTx(ctx, tx, &u)
			if err != nil {
				err = fmt.Errorf("error creating owner: %w", err)
				return
			}
		}
		if err != nil {
			err = fmt.Errorf("error getting owner's ID: %w", err)
			return
		}

		u.ID = userID
		results = append(results, u)
	}

	return
}

func (r *AssetRepository) execContext(ctx context.Context, execer sqlx.ExecerContext, query string, args ...interface{}) error {
	res, err := execer.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error running query: %w", err)
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %w", err)
	}
	if affectedRows == 0 {
		return errors.New("query affected 0 rows")
	}

	return nil
}

func (r *AssetRepository) buildSQL(builder sq.SelectBuilder) (query string, args []interface{}, err error) {
	query, args, err = builder.ToSql()
	if err != nil {
		err = fmt.Errorf("error transforming to sql")
		return
	}
	query, err = sq.Dollar.ReplacePlaceholders(query)
	if err != nil {
		err = fmt.Errorf("error replacing placeholders to dollar")
		return
	}

	return
}

func (r *AssetRepository) getAssetSQL() sq.SelectBuilder {
	return sq.Select(`
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
		u.email as "updated_by.email",
		u.provider as "updated_by.provider",
		u.created_at as "updated_by.created_at",
		u.updated_at as "updated_by.updated_at"
		`).
		From("assets a").
		LeftJoin("users u ON a.updated_by = u.id")
}

func (r *AssetRepository) getAssetVersionSQL() sq.SelectBuilder {
	return sq.Select(`
		a.asset_id as id,
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
		a.changelog as changelog,
		a.owners as owners,
		u.id as "updated_by.id",
		u.email as "updated_by.email",
		u.provider as "updated_by.provider",
		u.created_at as "updated_by.created_at",
		u.updated_at as "updated_by.updated_at"
		`).
		From("assets_versions a").
		LeftJoin("users u ON a.updated_by = u.id")
}

// NewAssetRepository initializes user repository clients
func NewAssetRepository(c *Client, userRepo *UserRepository, defaultGetMaxSize int, defaultUserProvider string) (*AssetRepository, error) {
	if c == nil {
		return nil, errors.New("postgres client is nil")
	}
	if defaultGetMaxSize == 0 {
		defaultGetMaxSize = DEFAULT_MAX_RESULT_SIZE
	}
	if defaultUserProvider == "" {
		defaultUserProvider = "unknown"
	}

	return &AssetRepository{
		client:              c,
		defaultGetMaxSize:   defaultGetMaxSize,
		defaultUserProvider: defaultUserProvider,
		userRepo:            userRepo,
	}, nil
}
