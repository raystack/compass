package worker

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrInvalidJobHandler = errors.New("job handler is not valid")

// Default run options
var (
	DefaultMaxAttempts                     = 3
	DefaultTimeout                         = 5 * time.Second
	DefaultBackoffStrategy BackoffStrategy = DefaultExponentialBackoff
)

// JobHandler is used to execute a job by the Worker when ready. The Handle
// function executes the given job with additional control via JobOpts.
type JobHandler struct {
	Handle  JobFunc
	JobOpts JobOptions
}

// JobFunc is invoked by the Worker when a job is ready. If it returns
// RetryableError, the Worker may retry the job execution with an appropriate
// backoff. If it returns any other error or if it panics, the job will be
// marked as a dead job.
type JobFunc func(context.Context, JobSpec) error

// JobOptions control the retry strategy and the job execution timeout.
type JobOptions struct {
	MaxAttempts int
	Timeout     time.Duration
	BackoffStrategy
}

// Sanitize sanitizes the job handler and sets defaults for unspecified job
// options.
// Returns ErrInvalidJobHandler if the Handle function is not set.
func (j *JobHandler) Sanitize() error {
	if j.Handle == nil {
		return fmt.Errorf("sanitize job handler: %w: handle function must be set", ErrInvalidJobHandler)
	}

	if j.JobOpts.MaxAttempts <= 0 {
		j.JobOpts.MaxAttempts = DefaultMaxAttempts
	}
	if j.JobOpts.Timeout <= 0 {
		j.JobOpts.Timeout = DefaultTimeout
	}
	if j.JobOpts.BackoffStrategy == nil {
		j.JobOpts.BackoffStrategy = DefaultBackoffStrategy
	}

	return nil
}
