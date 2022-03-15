package discussion

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidID    = errors.New("invalid discussion ID")
	ErrInvalidType  = fmt.Errorf("discussion type is invalid, supported types are: %s", strings.Join(SupportedTypes, ","))
	ErrInvalidState = fmt.Errorf("discussion state is invalid, supported states are: %s", strings.Join(SupportedStates, ","))
)

type NotFoundError struct {
	CommentID    string
	DiscussionID string
}

func (e NotFoundError) Error() string {
	reasons := []string{}
	if e.DiscussionID != "" {
		reasons = append(reasons, fmt.Sprintf("with discussion id \"%s\"", e.DiscussionID))
	}
	if e.CommentID != "" {
		reasons = append(reasons, fmt.Sprintf("with comment id \"%s\"", e.CommentID))
	}
	return "resource not found " + strings.Join(reasons, " and ")
}

type InvalidError struct {
	CommentID    string
	DiscussionID string
}

func (e InvalidError) Error() string {
	reasons := []string{}
	if e.DiscussionID != "" {
		reasons = append(reasons, fmt.Sprintf("with discussion id \"%s\"", e.DiscussionID))
	}
	if e.CommentID != "" {
		reasons = append(reasons, fmt.Sprintf("with comment id \"%s\"", e.CommentID))
	}
	return "invalid input " + strings.Join(reasons, " and ")
}
