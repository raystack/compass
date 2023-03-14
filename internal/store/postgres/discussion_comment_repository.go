package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/goto/compass/core/discussion"
)

// Create adds a new comment to a specific discussion
func (r *DiscussionRepository) CreateComment(ctx context.Context, cmt *discussion.Comment) (string, error) {
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

	if err = r.client.db.QueryRowContext(ctx, query, args...).Scan(&commentID); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, errForeignKeyViolation) {
			return "", discussion.NotFoundError{DiscussionID: cmt.DiscussionID}
		}

		return "", fmt.Errorf("error running insert query: %w", err)
	}

	if commentID == "" {
		return "", fmt.Errorf("error comment ID is empty from DB")
	}

	return commentID, nil
}

// GetAll fetches all comments of a specific discussion
func (r *DiscussionRepository) GetAllComments(ctx context.Context, did string, flt discussion.Filter) ([]discussion.Comment, error) {

	builder := r.selectCommentsSQL()
	builder = builder.Where(sq.Eq{"discussion_id": did})
	builder = r.buildSelectOrderQuery(builder, flt)
	builder = r.buildSelectLimitQuery(builder, flt)
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	cmts := []discussion.Comment{}
	err = r.client.db.SelectContext(ctx, &cmts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting list of comments: %w", err)
	}

	return cmts, nil
}

// Get fetchs a comment
func (r *DiscussionRepository) GetComment(ctx context.Context, cid string, did string) (discussion.Comment, error) {

	builder := r.selectCommentsSQL()
	builder = builder.Where(sq.Eq{
		"c.id":            cid,
		"c.discussion_id": did,
	})
	query, args, err := r.buildSQL(builder)
	if err != nil {
		return discussion.Comment{}, fmt.Errorf("error building query: %w", err)
	}

	cmt := discussion.Comment{}
	err = r.client.db.GetContext(ctx, &cmt, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return discussion.Comment{}, discussion.NotFoundError{CommentID: cid, DiscussionID: did}
	}
	if err != nil {
		return discussion.Comment{}, fmt.Errorf("error getting list of comments: %w", err)
	}

	return cmt, nil
}

// Update updates a comment
func (r *DiscussionRepository) UpdateComment(ctx context.Context, cmt *discussion.Comment) error {
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
		return discussion.NotFoundError{CommentID: cmt.ID, DiscussionID: cmt.DiscussionID}
	}

	return nil
}

// Delete removes a comment
func (r *DiscussionRepository) DeleteComment(ctx context.Context, cid string, did string) error {
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
		return discussion.NotFoundError{CommentID: cid, DiscussionID: did}
	}

	return nil
}

func (r *DiscussionRepository) selectCommentsSQL() sq.SelectBuilder {
	return sq.Select(`
		c.id as id,
		c.discussion_id as discussion_id,
		c.body as body,
		c.created_at as created_at,
		c.updated_at as updated_at,
		uo.id as "owner.id",
		uo.uuid as "owner.uuid",
		uo.email as "owner.email",
		uo.provider as "owner.provider",
		uo.created_at as "owner.created_at",
		uo.updated_at as "owner.updated_at",
		uu.uuid as "updated_by.uuid",
		uu.email as "updated_by.email",
		uu.provider as "updated_by.provider",
		uu.created_at as "updated_by.created_at",
		uu.updated_at as "updated_by.updated_at"
		`).
		From("comments c").
		Join("users uo ON c.owner = uo.id").
		Join("users uu ON c.updated_by = uu.id")
}
