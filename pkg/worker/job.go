package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

type JobStatus string

const (
	StatusUnknown = ""
	StatusDone    = "done"
	StatusDead    = "dead"
)

const minRetryBackoff = 5 * time.Second

var ErrInvalidJob = errors.New("job is not valid")

// JobSpec is the specification for async processing.
type JobSpec struct {
	Type    string    `json:"type"`
	Payload []byte    `json:"args"`
	RunAt   time.Time `json:"run_at"`
}

// Job represents the specification for async processing and also
// maintains the progress so far.
type Job struct {
	// Specification of the job.
	ID ulid.ULID `json:"id"`
	JobSpec

	// Internal metadata.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Execution information.
	AttemptsDone  int       `json:"attempts_done"`
	Status        JobStatus `json:"-"`
	LastAttemptAt time.Time `json:"last_attempt_at,omitempty"`
	LastError     string    `json:"last_error,omitempty"`
}

// NewJob sanitizes the given JobSpec and returns a new instance of
// Job created with the given job.
// Returns ErrInvalidJob if the job type is empty.
func NewJob(j JobSpec) (Job, error) {
	if j.Type == "" {
		return Job{}, fmt.Errorf("%w: job type must be set", ErrInvalidJob)
	}

	now := time.Now()

	j.Type = strings.TrimSpace(strings.ToLower(j.Type))
	if j.RunAt.IsZero() {
		j.RunAt = now
	}
	return Job{
		ID:        ulid.Make(),
		JobSpec:   j,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Attempt attempts to safely invoke the handler for this job. Handles success,
// failure and panic scenarios and updates the job with result in-place.
func (j *Job) Attempt(baseCtx context.Context, now time.Time, h JobHandler) {
	defer func() {
		if v := recover(); v != nil {
			j.LastError = fmt.Sprintf("panic: %v", v)
			j.Status = StatusDead
		}

		j.AttemptsDone++
		j.LastAttemptAt = now
		j.UpdatedAt = now
	}()

	select {
	case <-baseCtx.Done():
		j.RunAt = now.Add(minRetryBackoff)
		j.LastError = fmt.Sprintf("canceled: %v", baseCtx.Err())
		return

	default:
	}

	ctx, cancel := context.WithTimeout(baseCtx, h.JobOpts.Timeout)
	defer cancel()
	if err := h.Handle(ctx, j.JobSpec); err != nil {
		var re *RetryableError
		if errors.As(err, &re) && j.AttemptsDone+1 < h.JobOpts.MaxAttempts {
			j.RunAt = now.Add(h.JobOpts.Backoff(j.AttemptsDone + 1))
		} else {
			j.Status = StatusDead
		}
		j.LastError = err.Error()
		return
	}

	j.Status = StatusDone
}
