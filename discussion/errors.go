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
	DiscussionID string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("could not find discussion with id: %s", e.DiscussionID)
}
