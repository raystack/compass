package discussion_test

import (
	"fmt"
	"testing"

	"github.com/odpf/compass/core/discussion"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	comment := discussion.Comment{}
	t.Run("throws error for empty comment body", func(t *testing.T) {
		err := comment.Validate()
		assert.Equal(t, err, fmt.Errorf("body cannot be empty"))
	})
	t.Run("throws error for empty discussion id", func(t *testing.T) {
		comment.Body = "compass: sample body for comments"
		err := comment.Validate()
		assert.Equal(t, err, fmt.Errorf("discussion_id cannot be empty"))
	})
	t.Run("throws no error for correct comment", func(t *testing.T) {
		comment.Body = "compass: sample body for comments"
		comment.DiscussionID = "hit_1"
		err := comment.Validate()
		assert.NoError(t, err)
	})
}
