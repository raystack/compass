package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/star"
	"github.com/raystack/compass/core/user"
)

type StarClauses struct {
	Limit            int
	Offset           int
	SortKey          string
	SortDirectionKey string
}

type StarRepository struct {
	client *Client
}

func (r *StarRepository) Create(ctx context.Context, ns *namespace.Namespace, userID string, entityID string) (string, error) {
	var starID string
	if userID == "" {
		return "", star.ErrEmptyUserID
	}
	if entityID == "" {
		return "", star.ErrEmptyEntityID
	}
	if !isValidUUID(userID) {
		return "", star.InvalidError{UserID: userID}
	}
	if !isValidUUID(entityID) {
		return "", star.InvalidError{EntityID: entityID}
	}

	err := r.client.QueryFn(ctx, func(conn *sqlx.Conn) error {
		return conn.QueryRowxContext(ctx,
			`INSERT INTO stars (user_id, entity_id, namespace_id) VALUES ($1, $2, $3) RETURNING id`,
			userID, entityID, ns.ID).Scan(&starID)
	})
	if err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, errDuplicateKey) {
			return "", star.DuplicateRecordError{UserID: userID, EntityID: entityID}
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

func (r *StarRepository) GetStargazers(ctx context.Context, flt star.Filter, entityID string) ([]user.User, error) {
	if entityID == "" {
		return nil, star.ErrEmptyEntityID
	}
	if !isValidUUID(entityID) {
		return nil, star.InvalidError{EntityID: entityID}
	}

	clauses := r.buildClausesValue(flt)
	var userModels UserModels
	if err := r.client.SelectContext(ctx, &userModels, `
		SELECT DISTINCT ON (u.id) u.id, u.uuid, u.email, u.provider, u.created_at, u.updated_at
		FROM stars s
		JOIN users u ON s.user_id = u.id
		WHERE s.entity_id = $1
		LIMIT $2 OFFSET $3
	`, entityID, clauses.Limit, clauses.Offset); err != nil {
		return nil, fmt.Errorf("failed fetching stargazers: %w", err)
	}
	if len(userModels) == 0 {
		return nil, star.NotFoundError{EntityID: entityID}
	}
	return userModels.toUsers(), nil
}

func (r *StarRepository) GetAllEntitiesByUserID(ctx context.Context, flt star.Filter, userID string) ([]entity.Entity, error) {
	if userID == "" {
		return nil, star.ErrEmptyUserID
	}
	if !isValidUUID(userID) {
		return nil, star.InvalidError{UserID: userID}
	}

	clauses := r.buildClausesValue(flt)
	var models []entityModel
	if err := r.client.SelectContext(ctx, &models, fmt.Sprintf(`
		SELECT
			a.id, a.namespace_id, a.urn, a.type, a.name, a.description,
			a.properties, COALESCE(a.source, '') as source,
			a.valid_from, a.valid_to, a.created_at, a.updated_at
		FROM stars s
		INNER JOIN entities a ON s.entity_id = a.id
		WHERE s.user_id = $1 AND a.valid_to IS NULL
		ORDER BY $2 %s
		LIMIT $3 OFFSET $4
	`, clauses.SortDirectionKey), userID, clauses.SortKey, clauses.Limit, clauses.Offset); err != nil {
		return nil, fmt.Errorf("failed fetching starred entities: %w", err)
	}
	if len(models) == 0 {
		return nil, star.NotFoundError{UserID: userID}
	}

	result := make([]entity.Entity, len(models))
	for i, m := range models {
		result[i] = m.toEntity()
	}
	return result, nil
}

func (r *StarRepository) GetEntityByUserID(ctx context.Context, userID string, entityID string) (entity.Entity, error) {
	if userID == "" {
		return entity.Entity{}, star.ErrEmptyUserID
	}
	if entityID == "" {
		return entity.Entity{}, star.ErrEmptyEntityID
	}
	if !isValidUUID(userID) {
		return entity.Entity{}, star.InvalidError{UserID: userID}
	}
	if !isValidUUID(entityID) {
		return entity.Entity{}, star.InvalidError{EntityID: entityID}
	}

	var m entityModel
	err := r.client.GetContext(ctx, &m, `
		SELECT
			a.id, a.namespace_id, a.urn, a.type, a.name, a.description,
			a.properties, COALESCE(a.source, '') as source,
			a.valid_from, a.valid_to, a.created_at, a.updated_at
		FROM stars s
		INNER JOIN entities a ON s.entity_id = a.id
		WHERE s.user_id = $1 AND s.entity_id = $2 AND a.valid_to IS NULL
		LIMIT 1
	`, userID, entityID)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Entity{}, star.NotFoundError{EntityID: entityID, UserID: userID}
	}
	if err != nil {
		return entity.Entity{}, fmt.Errorf("failed fetching starred entity: %w", err)
	}
	return m.toEntity(), nil
}

func (r *StarRepository) Delete(ctx context.Context, userID string, entityID string) error {
	if userID == "" {
		return star.ErrEmptyUserID
	}
	if entityID == "" {
		return star.ErrEmptyEntityID
	}
	if !isValidUUID(userID) {
		return star.InvalidError{UserID: userID}
	}
	if !isValidUUID(entityID) {
		return star.InvalidError{EntityID: entityID}
	}

	res, err := r.client.ExecContext(ctx, `DELETE FROM stars WHERE user_id = $1 AND entity_id = $2`, userID, entityID)
	if err != nil {
		return fmt.Errorf("failed to unstar entity: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return star.NotFoundError{EntityID: entityID, UserID: userID}
	}
	return nil
}

func (r *StarRepository) buildClausesValue(flt star.Filter) StarClauses {
	sCfg := StarClauses{
		Offset:           0,
		Limit:            DefaultMaxResultSize,
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

func NewStarRepository(c *Client) (*StarRepository, error) {
	if c == nil {
		return nil, errNilPostgresClient
	}
	return &StarRepository{client: c}, nil
}
