package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/odpf/columbus/discussion"
)

// DiscussionRepository is a type that manages discussion operation to the primary database
type DiscussionRepository struct {
	client            *Client
	defaultGetMaxSize int
}

// GetAll fetchs all discussion data
func (r *DiscussionRepository) GetAll(ctx context.Context, flt discussion.Filter) ([]discussion.Discussion, error) {

	builder := r.selectSQL()
	builder = r.buildSelectFilterQuery(builder, flt)
	builder = r.buildSelectOrderQuery(builder, flt)
	builder = r.buildSelectLimitQuery(builder, flt)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	dms := []DiscussionModel{}
	err = r.client.db.SelectContext(ctx, &dms, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting discussion list: %w", err)
	}

	discussions := []discussion.Discussion{}
	for _, dm := range dms {
		discussions = append(discussions, dm.toDiscussion())
	}

	return discussions, nil
}

// Create inserts a new discussion data
func (r *DiscussionRepository) Create(ctx context.Context, dsc *discussion.Discussion) (string, error) {
	dm := newDiscussionModel(dsc)
	var discussionID string
	query, args, err := sq.Insert("discussions").
		Columns("title",
			"body",
			"state",
			"type",
			"owner",
			"labels",
			"assets",
			"assignees").
		Values(dm.Title, dm.Body, discussion.StateOpen, dm.Type, dm.Owner.ID, dm.Labels, dm.Assets, dm.Assignees).
		Suffix("RETURNING \"id\"").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("error building insert query: %w", err)
	}

	err = r.client.db.QueryRowContext(ctx, query, args...).Scan(&discussionID)
	if err != nil {
		return "", fmt.Errorf("error running insert query: %w", err)
	}

	if discussionID == "" {
		return "", fmt.Errorf("error discussion ID is empty from DB")
	}

	return discussionID, nil
}

// Get returns a specific discussion by id
func (r *DiscussionRepository) Get(ctx context.Context, did string) (discussion.Discussion, error) {
	builder := r.selectSQL()
	builder = builder.Where("d.id = ?", did).Limit(1)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return discussion.Discussion{}, fmt.Errorf("error building query: %w", err)
	}

	var discussionModel DiscussionModel
	err = r.client.db.GetContext(ctx, &discussionModel, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return discussion.Discussion{}, discussion.NotFoundError{DiscussionID: did}
	}
	if err != nil {
		return discussion.Discussion{}, fmt.Errorf("failed fetching discussion with id %s: %w", did, err)
	}

	return discussionModel.toDiscussion(), nil
}

// Patch will update a field in discussion
func (r *DiscussionRepository) Patch(ctx context.Context, dsc *discussion.Discussion) error {
	if dsc.ID == "" {
		return discussion.ErrInvalidID
	}

	dm := newDiscussionModel(dsc)
	builder := r.patchSQL(dm)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	res, err := r.client.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed updating discussion with id %s: %w", dm.ID, err)
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %w", err)
	}
	if affectedRows == 0 {
		return discussion.NotFoundError{DiscussionID: dm.ID}
	}

	return nil
}

func (r *DiscussionRepository) patchSQL(dm *DiscussionModel) sq.UpdateBuilder {
	builder := sq.Update("discussions")

	if len(strings.TrimSpace(dm.Title)) > 0 {
		builder = builder.Set("title", dm.Title)
	}

	if len(strings.TrimSpace(dm.Body)) > 0 {
		builder = builder.Set("body", dm.Body)
	}

	if len(strings.TrimSpace(dm.Type)) > 0 {
		builder = builder.Set("type", dm.Type)
	}

	if len(strings.TrimSpace(dm.State)) > 0 {
		builder = builder.Set("state", dm.State)
	}

	if dm.Labels != nil {
		if len(dm.Labels) > 0 {
			builder = builder.Set("labels", dm.Labels)
		} else {
			builder = builder.Set("labels", nil)
		}
	}

	if dm.Assets != nil {
		if len(dm.Assets) > 0 {
			builder = builder.Set("assets", dm.Assets)
		} else {
			builder = builder.Set("assets", nil)
		}
	}

	if dm.Assignees != nil {
		if len(dm.Assignees) > 0 {
			builder = builder.Set("assignees", dm.Assignees)
		} else {
			builder = builder.Set("assignees", nil)
		}
	}

	return builder.Where(sq.Eq{"id": dm.ID})
}

func (r *DiscussionRepository) selectSQL() sq.SelectBuilder {
	return sq.Select(`
		d.id as id,
		d.title as title,
		d.body as body,
		d.state as state,
		d.type as type,
		d.labels as labels,
		d.assets as assets,
		d.assignees as assignees,
		d.created_at as created_at,
		d.updated_at as updated_at,
		u.id as "owner.id",
		u.email as "owner.email",
		u.provider as "owner.provider",
		u.created_at as "owner.created_at",
		u.updated_at as "owner.updated_at"
		`).
		From("discussions d").
		LeftJoin("users u ON d.owner = u.id")
}

func (r *DiscussionRepository) buildSelectFilterQuery(builder sq.SelectBuilder, flt discussion.Filter) sq.SelectBuilder {
	whereClause := sq.Eq{}
	if flt.Type != "" {
		dTypeEnum := discussion.GetTypeEnum(flt.Type)
		whereClause["type"] = dTypeEnum
		if flt.State != "" {
			whereClause["state"] = discussion.GetStateEnum(flt.State)
		}
	}

	if flt.Owner != "" {
		whereClause["owner"] = flt.Owner
	}

	if len(whereClause) > 0 {
		builder = builder.Where(whereClause)
	}

	if len(flt.Labels) > 0 {
		builder = builder.Where("labels @> ?", flt.Labels)
	}

	if len(flt.Assignees) > 0 {
		builder = builder.Where("assignees @> ?", flt.Assignees)
	}

	if len(flt.Assets) > 0 {
		builder = builder.Where("assets @> ?", flt.Assets)
	}

	return builder
}

func (r *DiscussionRepository) buildSelectOrderQuery(builder sq.SelectBuilder, flt discussion.Filter) sq.SelectBuilder {
	if flt.SortBy != "" {
		orderDirection := "DESC"
		if flt.SortDirection != "" {
			orderDirection = strings.ToUpper(flt.SortDirection)
		}
		return builder.OrderBy(flt.SortBy + " " + orderDirection)
	}

	return builder
}

func (r *DiscussionRepository) buildSelectLimitQuery(builder sq.SelectBuilder, flt discussion.Filter) sq.SelectBuilder {
	limitSize := r.defaultGetMaxSize
	if flt.Size > 0 {
		limitSize = flt.Size
	}

	return builder.
		Limit(uint64(limitSize)).
		Offset(uint64(flt.Offset))
}

func (r *DiscussionRepository) buildSQL(builder sq.Sqlizer) (query string, args []interface{}, err error) {
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

// NewDiscussionRepository initializes discussion repository clients
func NewDiscussionRepository(c *Client, defaultGetMaxSize int) (*DiscussionRepository, error) {
	if c == nil {
		return nil, errors.New("postgres client is nil")
	}
	if defaultGetMaxSize == 0 {
		defaultGetMaxSize = DEFAULT_MAX_RESULT_SIZE
	}

	return &DiscussionRepository{
		client:            c,
		defaultGetMaxSize: defaultGetMaxSize,
	}, nil
}
