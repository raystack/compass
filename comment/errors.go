package comment

import (
	"fmt"
	"strings"
)

type NotFoundError struct {
	CommentID    string
	DiscussionID string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("could not find comment with %s in discussion id %s", e.CommentID, e.DiscussionID)
}

type InvalidError struct {
	CommentID    string
	DiscussionID string
}

func (e InvalidError) Error() string {
	fields := []string{"invalid"}
	if e.DiscussionID != "" {
		fields = append(fields, fmt.Sprintf("discussion id \"%s\"", e.DiscussionID))
	}
	if e.CommentID != "" {
		fields = append(fields, fmt.Sprintf("comment id \"%s\"", e.CommentID))
	}
	return strings.Join(fields, " ")
}
