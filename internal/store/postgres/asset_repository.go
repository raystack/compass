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
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/jmoiron/sqlx"
	"github.com/r3labs/diff/v2"
)

// AssetRepository is a type that manages user operation to the primary database
type AssetRepository struct {
	client              *Client
	userRepo            *UserRepository
	defaultGetMaxSize   int
	defaultUserProvider string
}

// GetAll retrieves list of assets with filters
func (r *AssetRepository) GetAll(ctx context.Context, flt asset.Filter) ([]asset.Asset, error) {
	builder := r.getAssetSQL().Offset(uint64(flt.Offset))
	size := flt.Size

	if size > 0 {
		builder = r.getAssetSQL().Limit(uint64(size)).Offset(uint64(flt.Offset))
	}
	builder = r.BuildFilterQuery(builder, flt)
	builder = r.buildOrderQuery(builder, flt)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	var ams []*AssetModel
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

// GetTypes fetches types with assets count for all available types
// and returns them as a map[typeName]count
func (r *AssetRepository) GetTypes(ctx context.Context, flt asset.Filter) (map[asset.Type]int, error) {

	builder := r.getAssetsGroupByCountSQL("type")
	builder = r.BuildFilterQuery(builder, flt)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return nil, fmt.Errorf("error building get type query: %w", err)
	}

	results := make(map[asset.Type]int)
	rows, err := r.client.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting type of assets: %w", err)
	}
	for rows.Next() {
		row := make(map[string]interface{})
		err = rows.MapScan(row)
		if err != nil {
			return nil, err
		}
		typeStr, ok := row["type"].(string)
		if !ok {
			return nil, err
		}
		typeCount, ok := row["count"].(int64)
		if !ok {
			return nil, err
		}
		typeName := asset.Type(typeStr)
		if typeName.IsValid() {
			results[typeName] = int(typeCount)
		}
	}

	return results, nil
}

// GetCount retrieves number of assets for every type
func (r *AssetRepository) GetCount(ctx context.Context, flt asset.Filter) (total int, err error) {
	builder := sq.Select("count(1)").From("assets")
	builder = r.BuildFilterQuery(builder, flt)
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
func (r *AssetRepository) GetByID(ctx context.Context, id string) (asset.Asset, error) {
	if !isValidUUID(id) {
		return asset.Asset{}, asset.InvalidError{AssetID: id}
	}

	ast, err := r.getWithPredicate(ctx, sq.Eq{"a.id": id})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{AssetID: id}
	}
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error getting asset with ID = %q: %w", id, err)
	}

	return ast, nil
}

func (r *AssetRepository) GetByURN(ctx context.Context, urn string) (asset.Asset, error) {
	ast, err := r.getWithPredicate(ctx, sq.Eq{"a.urn": urn})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{URN: urn}
	}
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error getting asset with URN = %q: %w", urn, err)
	}

	return ast, nil
}

func (r *AssetRepository) getWithPredicate(ctx context.Context, pred sq.Eq) (asset.Asset, error) {
	builder := r.getAssetSQL().
		Where(pred).
		Limit(1)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error building query: %w", err)
	}

	var am AssetModel
	err = r.client.db.GetContext(ctx, &am, query, args...)
	if err != nil {
		return asset.Asset{}, err
	}

	owners, err := r.getOwners(ctx, am.ID)
	if err != nil {
		return asset.Asset{}, err
	}

	return am.toAsset(owners), nil
}

// GetVersionHistory retrieves the versions of an asset
func (r *AssetRepository) GetVersionHistory(ctx context.Context, flt asset.Filter, id string) (avs []asset.Asset, err error) {
	if !isValidUUID(id) {
		err = asset.InvalidError{AssetID: id}
		return
	}

	size := flt.Size
	if size == 0 {
		size = r.defaultGetMaxSize
	}

	builder := r.getAssetVersionSQL().
		Where(sq.Eq{"a.asset_id": id}).
		OrderBy("string_to_array(version, '.')::int[] DESC").
		Limit(uint64(size)).
		Offset(uint64(flt.Offset))
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

// GetByVersionWithID retrieves the specific asset version
func (r *AssetRepository) GetByVersionWithID(ctx context.Context, id string, version string) (asset.Asset, error) {
	if !isValidUUID(id) {
		return asset.Asset{}, asset.InvalidError{AssetID: id}
	}

	ast, err := r.getByVersion(ctx, id, version, r.GetByID, sq.Eq{
		"a.asset_id": id,
		"a.version":  version,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{AssetID: id}
	}
	if err != nil {
		return asset.Asset{}, err
	}

	return ast, nil
}

func (r *AssetRepository) GetByVersionWithURN(ctx context.Context, urn string, version string) (asset.Asset, error) {
	ast, err := r.getByVersion(ctx, urn, version, r.GetByURN, sq.Eq{
		"a.urn":     urn,
		"a.version": version,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{URN: urn}
	}
	if err != nil {
		return asset.Asset{}, err
	}

	return ast, nil
}

type getAssetFunc func(context.Context, string) (asset.Asset, error)

func (r *AssetRepository) getByVersion(
	ctx context.Context, id, version string, get getAssetFunc, pred sq.Eq,
) (asset.Asset, error) {
	latest, err := get(ctx, id)
	if err != nil {
		return asset.Asset{}, err
	}

	if latest.Version == version {
		return latest, nil
	}

	var ast AssetModel
	builder := r.getAssetVersionSQL().
		Where(pred)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error building query: %w", err)
	}

	err = r.client.db.GetContext(ctx, &ast, query, args...)
	if err != nil {
		return asset.Asset{}, fmt.Errorf("failed fetching asset version: %w", err)
	}

	return ast.toVersionedAsset(latest)
}

// Upsert creates a new asset if it does not exist yet.
// It updates if asset does exist.
// Checking existence is done using "urn", "type", and "service" fields.
func (r *AssetRepository) Upsert(ctx context.Context, ast *asset.Asset) (string, error) {
	fetchedAsset, err := r.GetByURN(ctx, ast.URN)
	if errors.As(err, new(asset.NotFoundError)) {
		err = nil
	}
	if err != nil {
		return "", fmt.Errorf("error getting asset by URN: %w", err)
	}

	if fetchedAsset.ID == "" {
		// insert flow
		id, err := r.insert(ctx, ast)
		if err != nil {
			return "", fmt.Errorf("error inserting asset to DB: %w", err)
		}
		return id, nil
	}

	// update flow
	changelog, err := fetchedAsset.Diff(ast)
	if err != nil {
		return "", fmt.Errorf("error diffing two assets: %w", err)
	}

	err = r.update(ctx, fetchedAsset.ID, ast, &fetchedAsset, changelog)
	if err != nil {
		return "", fmt.Errorf("error updating asset to DB: %w", err)
	}

	return fetchedAsset.ID, nil
}

// DeleteByID removes asset using its ID
func (r *AssetRepository) DeleteByID(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return asset.InvalidError{AssetID: id}
	}

	affectedRows, err := r.deleteWithPredicate(ctx, sq.Eq{"id": id})
	if err != nil {
		return fmt.Errorf("error deleting asset with ID = %q: %w", id, err)
	}
	if affectedRows == 0 {
		return asset.NotFoundError{AssetID: id}
	}

	return nil
}

func (r *AssetRepository) DeleteByURN(ctx context.Context, urn string) error {
	affectedRows, err := r.deleteWithPredicate(ctx, sq.Eq{"urn": urn})
	if err != nil {
		return fmt.Errorf("error deleting asset with URN = %q: %w", urn, err)
	}
	if affectedRows == 0 {
		return asset.NotFoundError{URN: urn}
	}

	return nil
}

func (r *AssetRepository) AddProbe(ctx context.Context, assetURN string, probe *asset.Probe) error {
	probe.AssetURN = assetURN
	probe.CreatedAt = time.Now().UTC()
	if probe.Timestamp.IsZero() {
		probe.Timestamp = probe.CreatedAt
	} else {
		probe.Timestamp = probe.Timestamp.UTC()
	}

	query, args, err := sq.Insert("asset_probes").
		Columns("asset_urn", "status", "status_reason", "metadata", "timestamp", "created_at").
		Values(assetURN, probe.Status, probe.StatusReason, probe.Metadata, probe.Timestamp, probe.CreatedAt).
		Suffix("RETURNING \"id\"").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building insert asset probe query: %w", err)
	}

	err = r.client.db.QueryRowContext(ctx, query, args...).Scan(&probe.ID)
	if errors.Is(checkPostgresError(err), errForeignKeyViolation) {
		return asset.NotFoundError{URN: assetURN}
	} else if err != nil {
		return fmt.Errorf("error running insert asset probe query: %w", err)
	}

	return nil
}

func (r *AssetRepository) GetProbes(ctx context.Context, assetURN string) ([]asset.Probe, error) {
	query, args, err := sq.Select(
		"id", "asset_urn", "status", "status_reason", "metadata", "timestamp", "created_at",
	).From("asset_probes").
		OrderBy("created_at").
		Where(sq.Eq{"asset_urn": assetURN}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building get asset probes query: %w", err)
	}

	var models []AssetProbeModel
	if err := r.client.db.SelectContext(ctx, &models, query, args...); err != nil {
		return nil, fmt.Errorf("error running get asset probes query: %w", err)
	}

	results := []asset.Probe{}
	for _, m := range models {
		results = append(results, m.toAssetProbe())
	}

	return results, nil
}

func (r *AssetRepository) GetProbesWithFilter(ctx context.Context, flt asset.ProbesFilter) (map[string][]asset.Probe, error) {
	stmt := sq.Select(
		"id", "asset_urn", "status", "status_reason", "metadata", "timestamp", "created_at",
	).From("asset_probes").
		OrderBy("asset_urn", "timestamp DESC")

	if len(flt.AssetURNs) > 0 {
		stmt = stmt.Where(sq.Eq{"asset_urn": flt.AssetURNs})
	}
	if !flt.NewerThan.IsZero() {
		stmt = stmt.Where(sq.GtOrEq{"timestamp": flt.NewerThan})
	}
	if !flt.OlderThan.IsZero() {
		stmt = stmt.Where(sq.LtOrEq{"timestamp": flt.OlderThan})
	}
	if flt.MaxRows > 0 {
		stmt = stmt.Column("RANK() OVER (PARTITION BY asset_urn ORDER BY timestamp desc) rank_number")
		stmt = sq.Select(
			"id", "asset_urn", "status", "status_reason", "metadata", "timestamp", "created_at",
		).FromSelect(stmt, "ap").
			Where(sq.LtOrEq{"rank_number": flt.MaxRows})
	}

	query, args, err := stmt.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("get probes with filter: build query: %w", err)
	}

	var probes []AssetProbeModel
	if err := r.client.db.SelectContext(ctx, &probes, query, args...); err != nil {
		return nil, fmt.Errorf("error running get asset probes query: %w", err)
	}

	results := make(map[string][]asset.Probe, len(probes))
	for _, p := range probes {
		results[p.AssetURN] = append(results[p.AssetURN], p.toAssetProbe())
	}

	return results, nil
}

func (r *AssetRepository) deleteWithPredicate(ctx context.Context, pred sq.Eq) (int64, error) {
	query, args, err := r.buildSQL(sq.Delete("assets").Where(pred))
	if err != nil {
		return 0, fmt.Errorf("error building query: %w", err)
	}

	res, err := r.client.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting affected rows: %w", err)
	}

	return affectedRows, nil
}

func (r *AssetRepository) insert(ctx context.Context, ast *asset.Asset) (id string, err error) {
	err = r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		ast.CreatedAt = time.Now()
		ast.UpdatedAt = ast.CreatedAt
		query, args, err := sq.Insert("assets").
			Columns("urn", "type", "service", "name", "description", "data", "url", "labels", "created_at", "updated_by", "updated_at", "version").
			Values(ast.URN, ast.Type, ast.Service, ast.Name, ast.Description, ast.Data, ast.URL, ast.Labels, ast.CreatedAt, ast.UpdatedBy.ID, ast.UpdatedAt, asset.BaseVersion).
			Suffix("RETURNING \"id\"").
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return fmt.Errorf("error building insert query: %w", err)
		}

		ast.Version = asset.BaseVersion

		err = tx.QueryRowContext(ctx, query, args...).Scan(&id)
		if err != nil {
			return fmt.Errorf("error running insert query: %w", err)
		}

		users, err := r.createOrFetchUsers(ctx, tx, ast.Owners)
		if err != nil {
			return fmt.Errorf("error creating and fetching owners: %w", err)
		}

		err = r.insertOwners(ctx, tx, id, users)
		if err != nil {
			return fmt.Errorf("error running insert owners query: %w", err)
		}

		// insert versions
		ast.ID = id
		if err = r.insertAssetVersion(ctx, tx, ast, diff.Changelog{}); err != nil {
			return err
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
		newAsset.Version = newVersion
		newAsset.ID = oldAsset.ID
		newAsset.UpdatedAt = time.Now()

		query, args, err := r.buildSQL(sq.Update("assets").
			Set("urn", newAsset.URN).
			Set("type", newAsset.Type).
			Set("service", newAsset.Service).
			Set("name", newAsset.Name).
			Set("description", newAsset.Description).
			Set("data", newAsset.Data).
			Set("url", newAsset.URL).
			Set("labels", newAsset.Labels).
			Set("updated_at", newAsset.UpdatedAt).
			Set("updated_by", newAsset.UpdatedBy.ID).
			Set("version", newAsset.Version).
			Where(sq.Eq{"id": assetID}))
		if err != nil {
			return fmt.Errorf("build query: %w", err)
		}

		if err := r.execContext(ctx, tx, query, args...); err != nil {
			return fmt.Errorf("error running update asset query: %w", err)
		}

		// insert versions
		if err = r.insertAssetVersion(ctx, tx, newAsset, clog); err != nil {
			return err
		}

		// managing owners
		newAssetOwners, err := r.createOrFetchUsers(ctx, tx, newAsset.Owners)
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

	var userModels UserModels

	query := `
		SELECT
			u.id as "id",
			u.uuid as "uuid",
			u.email as "email",
			u.provider as "provider"
		FROM asset_owners ao
		JOIN users u on ao.user_id = u.id
		WHERE asset_id = $1`

	err = r.client.db.SelectContext(ctx, &userModels, query, assetID)
	if err != nil {
		err = fmt.Errorf("error getting asset's owners: %w", err)
	}

	owners = userModels.toUsers()

	return
}

// insertOwners inserts relation of asset id and user id
func (r *AssetRepository) insertOwners(ctx context.Context, execer sqlx.ExecerContext, assetID string, owners []user.User) error {
	if len(owners) == 0 {
		return nil
	}

	if !isValidUUID(assetID) {
		return asset.InvalidError{AssetID: assetID}
	}

	sqlb := sq.Insert("asset_owners").
		Columns("asset_id", "user_id")
	for _, o := range owners {
		sqlb = sqlb.Values(assetID, o.ID)
	}

	qry, args, err := sqlb.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build insert owners SQL: %w", err)
	}

	if err := r.execContext(ctx, execer, qry, args...); err != nil {
		return fmt.Errorf("error running insert owners query: %w", err)
	}

	return nil
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

func (r *AssetRepository) createOrFetchUsers(ctx context.Context, tx *sqlx.Tx, users []user.User) ([]user.User, error) {
	ids := make(map[string]struct{}, len(users))
	var results []user.User
	for _, u := range users {
		if u.ID != "" {
			if _, ok := ids[u.ID]; ok {
				continue
			}
			ids[u.ID] = struct{}{}
			results = append(results, u)
			continue
		}

		var (
			userID      string
			fetchedUser user.User
			err         error
		)
		if u.UUID != "" {
			fetchedUser, err = r.userRepo.GetByUUIDWithTx(ctx, tx, u.UUID)
		} else {
			fetchedUser, err = r.userRepo.GetByEmailWithTx(ctx, tx, u.Email)
		}
		switch {
		case errors.As(err, &user.NotFoundError{}):
			u.Provider = r.defaultUserProvider
			userID, err = r.userRepo.CreateWithTx(ctx, tx, &u)
			if err != nil {
				return nil, fmt.Errorf("error creating owner: %w", err)
			}

		case err != nil:
			return nil, fmt.Errorf("error getting owner's ID: %w", err)

		case err == nil:
			userID = fetchedUser.ID
		}

		if _, ok := ids[userID]; ok {
			continue
		}
		ids[userID] = struct{}{}
		u.ID = userID
		results = append(results, u)
	}

	return results, nil
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

type sqlBuilder interface {
	ToSql() (string, []interface{}, error)
}

func (r *AssetRepository) buildSQL(builder sqlBuilder) (query string, args []interface{}, err error) {
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

func (r *AssetRepository) getAssetsGroupByCountSQL(columnName string) sq.SelectBuilder {
	return sq.Select(columnName, "count(1)").
		From("assets").
		GroupBy(columnName)
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
		COALESCE(a.url, '') as url,
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
		u.uuid as "updated_by.uuid",
		u.email as "updated_by.email",
		u.provider as "updated_by.provider",
		u.created_at as "updated_by.created_at",
		u.updated_at as "updated_by.updated_at"
		`).
		From("assets_versions a").
		LeftJoin("users u ON a.updated_by = u.id")
}

// BuildFilterQuery retrieves the sql query based on applied filter in the queryString
func (r *AssetRepository) BuildFilterQuery(builder sq.SelectBuilder, flt asset.Filter) sq.SelectBuilder {
	if len(flt.Types) > 0 {
		builder = builder.Where(sq.Eq{"type": flt.Types})
	}

	if len(flt.Services) > 0 {
		builder = builder.Where(sq.Eq{"service": flt.Services})
	}

	if len(flt.QueryFields) > 0 && flt.Query != "" {
		orClause := sq.Or{}

		for _, field := range flt.QueryFields {
			finalQuery := field

			if strings.Contains(field, "data") {
				finalQuery = r.buildDataField(
					strings.TrimPrefix(field, "data."),
					false,
				)
			}
			orClause = append(orClause, sq.ILike{
				finalQuery: fmt.Sprint("%", flt.Query, "%"),
			})
		}
		builder = builder.Where(orClause)
	}

	if len(flt.Data) > 0 {
		for key, vals := range flt.Data {
			if len(vals) == 1 && vals[0] == "_nonempty" {
				field := r.buildDataField(key, true)
				whereClause := sq.And{
					sq.NotEq{field: nil},    // IS NOT NULL (field exists)
					sq.NotEq{field: "null"}, // field is not "null" JSON
					sq.NotEq{field: "[]"},   // field is not empty array
					sq.NotEq{field: "{}"},   // field is not empty object
					sq.NotEq{field: "\"\""}, // field is not empty string
				}
				builder = builder.Where(whereClause)
			} else {
				dataOrClause := sq.Or{}
				for _, v := range vals {
					finalQuery := r.buildDataField(key, false)
					dataOrClause = append(dataOrClause, sq.Eq{finalQuery: v})
				}

				builder = builder.Where(dataOrClause)
			}
		}
	}

	return builder
}

// buildFilterQuery retrieves the ordered sql query based on the sorting filter used in queryString
func (r *AssetRepository) buildOrderQuery(builder sq.SelectBuilder, flt asset.Filter) sq.SelectBuilder {
	if flt.SortBy == "" {
		return builder
	}

	orderDirection := "ASC"
	if flt.SortDirection != "" {
		orderDirection = flt.SortDirection
	}

	return builder.OrderBy(flt.SortBy + " " + orderDirection)
}

// buildDataField is a helper function to build nested data fields
func (r *AssetRepository) buildDataField(key string, asJsonB bool) (finalQuery string) {
	var queries []string

	queries = append(queries, "data")
	nestedParams := strings.Split(key, ".")
	totalParams := len(nestedParams)
	for i := 0; i < totalParams-1; i++ {
		nestedQuery := fmt.Sprintf("->'%s'", nestedParams[i])
		queries = append(queries, nestedQuery)
	}

	var lastParam string
	if asJsonB {
		lastParam = fmt.Sprintf("->'%s'", nestedParams[totalParams-1])
	} else {
		lastParam = fmt.Sprintf("->>'%s'", nestedParams[totalParams-1])
	}

	queries = append(queries, lastParam)
	finalQuery = strings.Join(queries, "")

	return finalQuery
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
