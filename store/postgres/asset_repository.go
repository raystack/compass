package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/user"
)

// AssetRepository is a type that manages user operation to the primary database
type AssetRepository struct {
	client            *Client
	userRepo          *UserRepository
	defaultGetMaxSize int
}

// Get retrieves list of assets with filters via config
func (r *AssetRepository) Get(ctx context.Context, config asset.Config) (assets []asset.Asset, err error) {
	size := config.Size
	if size == 0 {
		size = r.defaultGetMaxSize
	}

	builder := sq.Select("*").From("assets").
		Limit(uint64(size)).
		Offset(uint64(config.Offset))
	builder = r.buildFilterQuery(builder, config)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		err = fmt.Errorf("error building query: %w", err)
		return
	}

	ams := []*AssetModel{}
	err = r.client.db.SelectContext(ctx, &ams, query, args...)
	if err != nil {
		err = fmt.Errorf("error getting asset list: %w", err)
		return
	}

	assets = []asset.Asset{}
	for _, am := range ams {
		assets = append(assets, am.toAsset())
	}

	return
}

// Get retrieves list of assets with filters via config
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
		return asset.Asset{}, asset.InvalidError{AssetID: id}
	}

	query := `SELECT * FROM assets WHERE id = $1 LIMIT 1;`

	am := &AssetModel{}
	err = r.client.db.GetContext(ctx, am, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		err = asset.NotFoundError{AssetID: id}
		return
	}
	if err != nil {
		err = fmt.Errorf("error getting asset with ID = \"%s\": %w", id, err)
		return
	}
	ast = am.toAsset()

	owners, err := r.getOwners(ctx, id)
	if err != nil {
		err = fmt.Errorf("error getting asset's owners with ID = \"%s\": %w", id, err)
		return
	}
	ast.Owners = owners

	return
}

// Upsert creates a new asset if it does not exist yet.
// It updates if asset does exist.
// Checking existance is done using "urn", "type", and "service" fields.
func (r *AssetRepository) Upsert(ctx context.Context, ast *asset.Asset) (string, error) {
	fetchedAsset, err := r.getAssetByURN(ctx, ast.URN, ast.Type, ast.Service)
	if errors.As(err, new(asset.NotFoundError)) {
		err = nil
	}
	if err != nil {
		return "", fmt.Errorf("error getting asset ID: %w", err)
	}
	if fetchedAsset.ID == "" {
		fetchedAsset.ID, err = r.insert(ctx, ast)
		if err != nil {
			return fetchedAsset.ID, fmt.Errorf("error inserting asset to DB: %w", err)
		}
	} else {
		err = r.update(ctx, fetchedAsset.ID, ast)
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
		err := tx.QueryRowxContext(ctx,
			`INSERT INTO assets 
				(urn, type, service, name, description, data, labels, created_at, updated_at)
			VALUES 
				($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`,
			ast.URN, ast.Type, ast.Service, ast.Name, ast.Description, ast.Data, ast.Labels, ast.CreatedAt, ast.UpdatedAt).Scan(&id)
		if err != nil {
			return fmt.Errorf("error running insert query: %w", err)
		}

		ast.Owners, err = r.createOrFetchOwnersID(ctx, tx, ast.Owners)
		if err != nil {
			return fmt.Errorf("error creating and fetching owners: %w", err)
		}

		err = r.insertOwners(ctx, tx, id, ast.Owners)
		if err != nil {
			return fmt.Errorf("error running insert owners query: %w", err)
		}

		return nil
	})

	return
}

func (r *AssetRepository) update(ctx context.Context, id string, ast *asset.Asset) error {

	if !isValidUUID(id) {
		return asset.InvalidError{AssetID: id}
	}

	return r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		err := r.execContext(ctx, tx,
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
		if err != nil {
			return fmt.Errorf("error running update asset query: %w", err)
		}

		ast.Owners, err = r.createOrFetchOwnersID(ctx, tx, ast.Owners)
		if err != nil {
			return fmt.Errorf("error creating and fetching owners: %w", err)
		}
		currentOwners, err := r.getOwners(ctx, id)
		if err != nil {
			return fmt.Errorf("error getting asset's current owners: %w", err)
		}
		toInserts, toRemoves := r.compareOwners(currentOwners, ast.Owners)
		if err := r.insertOwners(ctx, tx, id, toInserts); err != nil {
			return fmt.Errorf("error inserting asset's new owners: %w", err)
		}
		if err := r.removeOwners(ctx, tx, id, toRemoves); err != nil {
			return fmt.Errorf("error removing asset's old owners: %w", err)
		}

		return nil
	})
}

// getOwners retrieves asset's owners by its ID
func (r *AssetRepository) getOwners(ctx context.Context, assetID string) (owners []user.User, err error) {

	if !isValidUUID(assetID) {
		return nil, asset.InvalidError{AssetID: assetID}
	}

	query := `
		SELECT u.id,u.email,u.provider
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

func (r *AssetRepository) createOrFetchOwnersID(ctx context.Context, tx *sqlx.Tx, users []user.User) (results []user.User, err error) {
	for _, u := range users {
		if u.ID != "" {
			continue
		}
		var userID string
		userID, err = r.userRepo.GetID(ctx, u.Email)
		if errors.As(err, &user.NotFoundError{}) {
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

func (r *AssetRepository) getAssetByURN(ctx context.Context, assetURN string, assetType asset.Type, assetService string) (asset.Asset, error) {
	query := `SELECT * FROM assets WHERE urn = $1 AND type = $2 AND service = $3;`
	var assetModel AssetModel
	if err := r.client.db.GetContext(ctx, &assetModel, query, assetURN, assetType, assetService); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return asset.Asset{}, asset.NotFoundError{}
		}
		return asset.Asset{}, fmt.Errorf(
			"error getting asset's ID with urn = \"%s\", type = \"%s\", service = \"%s\": %w",
			assetURN, assetType, assetService, err)
	}

	return assetModel.toAsset(), nil
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

func (r *AssetRepository) compareOwners(current, new []user.User) (toInserts, toRemove []user.User) {
	if len(current) == 0 && len(new) == 0 {
		return
	}

	currMap := map[string]int{}
	for _, curr := range current {
		currMap[curr.ID] = 1
	}

	for _, n := range new {
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

// NewAssetRepository initializes user repository clients
func NewAssetRepository(c *Client, userRepo *UserRepository, defaultGetMaxSize int) (*AssetRepository, error) {
	if c == nil {
		return nil, errors.New("postgres client is nil")
	}
	if defaultGetMaxSize == 0 {
		defaultGetMaxSize = DEFAULT_MAX_RESULT_SIZE
	}

	return &AssetRepository{
		client:            c,
		defaultGetMaxSize: defaultGetMaxSize,
		userRepo:          userRepo,
	}, nil
}
