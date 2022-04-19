package discussion

import (
	"fmt"
	"strings"
	"time"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Comment struct {
	ID           string    `json:"id" db:"id"`
	DiscussionID string    `json:"discussion_id" db:"discussion_id"`
	Body         string    `json:"body" db:"body"`
	Owner        user.User `json:"owner" db:"owner"`
	UpdatedBy    user.User `json:"updated_by" db:"updated_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ToProto transforms struct to proto
func (d Comment) ToProto() *compassv1beta1.Comment {

	var createdAtPB *timestamppb.Timestamp
	if !d.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(d.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !d.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(d.UpdatedAt)
	}

	return &compassv1beta1.Comment{
		Id:           d.ID,
		DiscussionId: d.DiscussionID,
		Body:         d.Body,
		Owner:        d.Owner.ToProto(),
		UpdatedBy:    d.UpdatedBy.ToProto(),
		CreatedAt:    createdAtPB,
		UpdatedAt:    updatedAtPB,
	}
}

// NewCommentFromProto transforms proto to struct
func NewCommentFromProto(pb *compassv1beta1.Comment) Comment {
	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	var owner user.User
	if pb.GetOwner() != nil {
		owner = user.NewFromProto(pb.GetOwner())
	}

	var updatedBy user.User
	if pb.GetUpdatedBy() != nil {
		updatedBy = user.NewFromProto(pb.GetUpdatedBy())
	}

	return Comment{
		ID:           pb.GetId(),
		DiscussionID: pb.GetDiscussionId(),
		Body:         pb.GetBody(),
		Owner:        owner,
		UpdatedBy:    updatedBy,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

// Validate checks emptyness required fields and constraint in comment and return error if the required is empty
func (c Comment) Validate() error {
	if len(strings.TrimSpace(c.Body)) == 0 {
		return fmt.Errorf("body cannot be empty")
	}

	if len(strings.TrimSpace(c.DiscussionID)) == 0 {
		return fmt.Errorf("discussion_id cannot be empty")
	}
	return nil
}
