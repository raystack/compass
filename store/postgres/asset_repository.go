package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/user"
)

const (
	DEFAULT_MAX_SIZE = 100
)

// AssetRepository is a type that manages user operation to the primary database
type AssetRepository struct {
	client            *Client
	userRepo          *UserRepository
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
		assets = append(assets, am.toAsset())
	}

	return
}

// Get retrieves list of assets with filters via config
func (r *AssetRepository) GetCount(ctx context.Context, config asset.GetConfig) (total int, err error) {
	query, args := r.buildGetCountQuery(config)
	err = r.client.db.GetContext(ctx, &total, query, args...)
	if err != nil {
		err = fmt.Errorf("error getting asset list: %w", err)
	}

	return
}

// GetByID retrieves asset by its ID
func (r *AssetRepository) GetByID(ctx context.Context, id string) (ast asset.Asset, err error) {
	query := `SELECT * FROM assets WHERE id = $1 LIMIT 1;`

	am := &Asset{}
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

func (r *AssetRepository) buildGetCountQuery(config asset.GetConfig) (query string, args []interface{}) {
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

	args = append(args)

	whereClauses := []string{}
	for i, field := range whereFields {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", field, i+1))
	}
	totalWhereClauses := len(whereClauses)

	query = "SELECT count(1) FROM assets "
	if totalWhereClauses > 0 {
		query += "WHERE " + strings.Join(whereClauses, " AND ") + " "
	}

	return
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
	return r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
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
func (r *AssetRepository) getOwners(ctx context.Context, asset_id string) (owners []user.User, err error) {
	query := `
		SELECT u.id,u.email,u.provider
		FROM asset_owners ao
		JOIN users u on ao.user_id = u.id
		WHERE asset_id = $1`
	ums := []User{}
	err = r.client.db.SelectContext(ctx, &ums, query, asset_id)
	if err != nil {
		err = fmt.Errorf("error getting asset's owners: %w", err)
	}
	for _, um := range ums {
		owners = append(owners, *um.toUser())
	}

	return
}

func (r *AssetRepository) insertOwners(ctx context.Context, execer sqlx.ExecerContext, asset_id string, owners []user.User) (err error) {
	if len(owners) == 0 {
		return
	}

	var values []string
	var args = []interface{}{asset_id}
	for i, owner := range owners {
		values = append(values, fmt.Sprintf("($1, $%d)", i+2))
		args = append(args, owner.ID)
	}
	query := fmt.Sprintf(`
		INSERT INTO asset_owners
			(asset_id, user_id)
		VALUES %s`, strings.Join(values, ","))
	_, err = execer.ExecContext(ctx, query, args...)
	if err != nil {
		err = fmt.Errorf("error running insert owners query: %w", err)
	}

	return
}

func (r *AssetRepository) removeOwners(ctx context.Context, execer sqlx.ExecerContext, asset_id string, owners []user.User) (err error) {
	if len(owners) == 0 {
		return
	}

	var user_ids []string
	var args = []interface{}{asset_id}
	for i, owner := range owners {
		user_ids = append(user_ids, fmt.Sprintf("$%d", i+2))
		args = append(args, owner.ID)
	}
	query := fmt.Sprintf(
		`DELETE FROM asset_owners WHERE asset_id = $1 AND user_id in (%s)`,
		strings.Join(user_ids, ","),
	)
	_, err = execer.ExecContext(ctx, query, args...)
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

	for id, _ := range currMap {
		toRemove = append(toRemove, user.User{ID: id})
	}

	return
}

// NewAssetRepository initializes user repository clients
func NewAssetRepository(c *Client, userRepo *UserRepository, defaultGetMaxSize int) (*AssetRepository, error) {
	if c == nil {
		return nil, errors.New("postgres client is nil")
	}
	if defaultGetMaxSize == 0 {
		defaultGetMaxSize = DEFAULT_MAX_SIZE
	}

	return &AssetRepository{
		client:            c,
		defaultGetMaxSize: defaultGetMaxSize,
		userRepo:          userRepo,
	}, nil
}
