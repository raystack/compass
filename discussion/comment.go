package discussion

import (
	"fmt"
	"strings"
	"time"

	"github.com/odpf/columbus/user"
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
