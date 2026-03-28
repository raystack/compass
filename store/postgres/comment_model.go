package postgres

import (
	"time"

	"github.com/raystack/compass/core/discussion"
)

type CommentModel struct {
	ID           string    `db:"id"`
	DiscussionID string    `db:"discussion_id"`
	Body         string    `db:"body"`
	Owner        UserModel `db:"owner"`
	UpdatedBy    UserModel `db:"updated_by"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (cm CommentModel) toComment() discussion.Comment {
	return discussion.Comment{
		ID:           cm.ID,
		DiscussionID: cm.DiscussionID,
		Body:         cm.Body,
		Owner:        cm.Owner.toUser(),
		UpdatedBy:    cm.UpdatedBy.toUser(),
		CreatedAt:    cm.CreatedAt,
		UpdatedAt:    cm.UpdatedAt,
	}
}
