package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/odpf/columbus/comment"
)

// CommentRepository is a type that manages comment operation to the primary database
type CommentRepository struct {
	client            *Client
	defaultGetMaxSize int
}

// Create adds a new comment to a specific discussion
func (r *CommentRepository) Create(ctx context.Context, cmt *comment.Comment) (string, error) {
	var commentID string
	query, args, err := sq.Insert("comments").
		Columns("discussion_id",
			"body",
			"owner",
			"updated_by").
		Values(cmt.DiscussionID, cmt.Body, cmt.Owner.ID, cmt.Owner.ID).
		Suffix("RETURNING \"id\"").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("error building insert query: %w", err)
	}

	err = r.client.db.QueryRowContext(ctx, query, args...).Scan(&commentID)
	if err != nil {
		return "", fmt.Errorf("error running insert query: %w", err)
	}

	if commentID == "" {
		return "", fmt.Errorf("error comment ID is empty from DB")
	}

	return commentID, nil
}

// GetAll fetchs all comments of a specific discussion
func (r *CommentRepository) GetAll(ctx context.Context, did string, flt comment.Filter) ([]comment.Comment, error) {

	builder := r.selectSQL()
	builder = builder.Where(sq.Eq{"discussion_id": did})
	builder = r.buildSelectOrderQuery(builder, flt)
	builder = r.buildSelectLimitQuery(builder, flt)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	cmts := []comment.Comment{}
	err = r.client.db.SelectContext(ctx, &cmts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting list of comments: %w", err)
	}

	return cmts, nil
}

// Get fetchs a comment
func (r *CommentRepository) Get(ctx context.Context, cid string, did string) (comment.Comment, error) {

	builder := r.selectSQL()
	builder = builder.Where(sq.Eq{
		"c.id":            cid,
		"c.discussion_id": did,
	})
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return comment.Comment{}, fmt.Errorf("error building query: %w", err)
	}

	cmt := comment.Comment{}
	err = r.client.db.GetContext(ctx, &cmt, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return comment.Comment{}, comment.NotFoundError{CommentID: cid, DiscussionID: did}
	}
	if err != nil {
		return comment.Comment{}, fmt.Errorf("error getting list of comments: %w", err)
	}

	return cmt, nil
}

// Update updates a comment
func (r *CommentRepository) Update(ctx context.Context, cmt *comment.Comment) error {
	builder := sq.Update("comments").
		Set("body", cmt.Body).
		Set("updated_by", cmt.UpdatedBy.ID).
		Set("updated_at", time.Now()).
		Where(sq.Eq{
			"id":            cmt.ID,
			"discussion_id": cmt.DiscussionID,
		})
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	res, err := r.client.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed updating comment with id %s and discussion id %s: %w", cmt.ID, cmt.DiscussionID, err)
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %w", err)
	}
	if affectedRows == 0 {
		return comment.NotFoundError{CommentID: cmt.ID, DiscussionID: cmt.DiscussionID}
	}

	return nil
}

// Delete removes a comment
func (r *CommentRepository) Delete(ctx context.Context, cid string, did string) error {
	builder := sq.Delete("comments").
		Where(sq.Eq{
			"id":            cid,
			"discussion_id": did,
		})
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	res, err := r.client.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed deleting comment with id %s and discussion id: %w", cid, err)
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %w", err)
	}
	if affectedRows == 0 {
		return comment.NotFoundError{CommentID: cid, DiscussionID: did}
	}

	return nil
}

func (r *CommentRepository) selectSQL() sq.SelectBuilder {
	return sq.Select(`
		c.id as id,
		c.discussion_id as discussion_id,
		c.body as body,
		c.created_at as created_at,
		c.updated_at as updated_at,
		uo.id as "owner.id",
		uo.email as "owner.email",
		uo.provider as "owner.provider",
		uo.created_at as "owner.created_at",
		uo.updated_at as "owner.updated_at",
		uu.id as "updated_by.id",
		uu.email as "updated_by.email",
		uu.provider as "updated_by.provider",
		uu.created_at as "updated_by.created_at",
		uu.updated_at as "updated_by.updated_at"
		`).
		From("comments c").
		Join("users uo ON c.owner = uo.id").
		Join("users uu ON c.updated_by = uu.id")
}

func (r *CommentRepository) buildSelectOrderQuery(builder sq.SelectBuilder, flt comment.Filter) sq.SelectBuilder {
	if flt.SortBy != "" {
		orderDirection := "DESC"
		if flt.SortDirection != "" {
			orderDirection = strings.ToUpper(flt.SortDirection)
		}
		return builder.OrderBy(flt.SortBy + " " + orderDirection)
	}

	return builder
}

func (r *CommentRepository) buildSelectLimitQuery(builder sq.SelectBuilder, flt comment.Filter) sq.SelectBuilder {
	limitSize := r.defaultGetMaxSize
	if flt.Size > 0 {
		limitSize = flt.Size
	}

	return builder.
		Limit(uint64(limitSize)).
		Offset(uint64(flt.Offset))
}

func (r *CommentRepository) buildSQL(builder sq.Sqlizer) (query string, args []interface{}, err error) {
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

// NewCommentRepository initializes comment repository clients
func NewCommentRepository(c *Client, defaultGetMaxSize int) (*CommentRepository, error) {
	if c == nil {
		return nil, errors.New("postgres client is nil")
	}
	if defaultGetMaxSize == 0 {
		defaultGetMaxSize = DEFAULT_MAX_RESULT_SIZE
	}

	return &CommentRepository{
		client:            c,
		defaultGetMaxSize: defaultGetMaxSize,
	}, nil
}
