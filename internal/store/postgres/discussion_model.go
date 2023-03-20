package postgres

import (
	"time"

	"github.com/lib/pq"
	"github.com/odpf/compass/core/discussion"
)

type DiscussionModel struct {
	ID          string         `db:"id"`
	NamespaceID string         `db:"namespace_id"`
	Title       string         `db:"title"`
	Body        string         `db:"body"`
	Type        string         `db:"type"`
	State       string         `db:"state"`
	Owner       UserModel      `db:"owner"`
	Labels      pq.StringArray `db:"labels"`
	Assets      pq.StringArray `db:"assets"`
	Assignees   pq.StringArray `db:"assignees"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func (dm DiscussionModel) toDiscussion() discussion.Discussion {
	return discussion.Discussion{
		ID:        dm.ID,
		Title:     dm.Title,
		Body:      dm.Body,
		Type:      discussion.GetTypeEnum(dm.Type),
		State:     discussion.GetStateEnum(dm.State),
		Labels:    dm.Labels,
		Assets:    dm.Assets,
		Assignees: dm.Assignees,
		Owner:     dm.Owner.toUser(),
		CreatedAt: dm.CreatedAt,
		UpdatedAt: dm.UpdatedAt,
	}
}

func newDiscussionModel(dc *discussion.Discussion) *DiscussionModel {
	um := newUserModel(&dc.Owner)
	return &DiscussionModel{
		ID:        dc.ID,
		Title:     dc.Title,
		Body:      dc.Body,
		Type:      dc.Type.String(),
		State:     dc.State.String(),
		Owner:     um,
		Labels:    dc.Labels,
		Assets:    dc.Assets,
		Assignees: dc.Assignees,
		CreatedAt: dc.CreatedAt,
		UpdatedAt: dc.UpdatedAt,
	}
}
