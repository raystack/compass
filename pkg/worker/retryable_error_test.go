package worker_test

import (
	"errors"
	"testing"

	"github.com/goto/compass/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func TestRetryableError_Error(t *testing.T) {
	err := &worker.RetryableError{Cause: errors.New("something smells fishy")}
	assert.Equal(t, "retryable-error: something smells fishy", err.Error())
}

func TestRetryableError_Unwrap(t *testing.T) {
	err := errors.New("something smells fishy")
	assert.ErrorIs(t, &worker.RetryableError{Cause: err}, err)
}
