package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/goto/compass/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func TestJob_Attempt(t *testing.T) {
	cancelledCtx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()

	createdAt := time.Unix(1654081526, 0)
	frozenTime := time.Unix(1654082526, 0)

	cases := []struct {
		name     string
		ctx      context.Context
		h        worker.JobHandler
		job      worker.Job
		expected worker.Job
	}{
		{
			name: "ContextCancelled",
			ctx:  cancelledCtx,
			job: worker.Job{
				UpdatedAt:     createdAt,
				LastAttemptAt: frozenTime,
			},
			h: worker.JobHandler{
				Handle: func(ctx context.Context, job worker.JobSpec) error {
					return nil
				},
				JobOpts: worker.JobOptions{MaxAttempts: 3, Timeout: time.Second},
			},
			expected: worker.Job{
				JobSpec: worker.JobSpec{
					RunAt: frozenTime.Add(5 * time.Second),
				},
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				LastError:     "canceled: context deadline exceeded",
			},
		},
		{
			name: "Panic",
			job: worker.Job{
				UpdatedAt:     createdAt,
				LastAttemptAt: frozenTime,
			},
			h: worker.JobHandler{
				Handle: func(ctx context.Context, job worker.JobSpec) error {
					panic("blown up")
				},
				JobOpts: worker.JobOptions{MaxAttempts: 3, Timeout: time.Second},
			},
			expected: worker.Job{
				Status:        worker.StatusDead,
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				LastError:     "panic: blown up",
			},
		},
		{
			name: "NonRetryableError",
			job: worker.Job{
				UpdatedAt:     createdAt,
				LastAttemptAt: frozenTime,
			},
			h: worker.JobHandler{
				Handle: func(ctx context.Context, job worker.JobSpec) error {
					return errors.New("a non-retryable error occurred")
				},
				JobOpts: worker.JobOptions{MaxAttempts: 3, Timeout: time.Second},
			},
			expected: worker.Job{
				Status:        worker.StatusDead,
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				LastError:     "a non-retryable error occurred",
			},
		},
		{
			name: "RetryableError",
			job: worker.Job{
				UpdatedAt:     createdAt,
				LastAttemptAt: frozenTime,
			},
			h: worker.JobHandler{
				Handle: func(ctx context.Context, job worker.JobSpec) error {
					return &worker.RetryableError{
						Cause: errors.New("some retryable error occurred"),
					}
				},
				JobOpts: worker.JobOptions{
					MaxAttempts: 3,
					Timeout:     time.Second,
					BackoffStrategy: worker.LinearBackoff{
						InitialDelay: 1 * time.Second,
						MaxDelay:     10 * time.Second,
					},
				},
			},
			expected: worker.Job{
				JobSpec: worker.JobSpec{
					RunAt: frozenTime.Add(1 * time.Second),
				},
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				LastError:     "retryable-error: some retryable error occurred",
			},
		},
		{
			name: "SuccessfulFirstAttempt",
			job: worker.Job{
				UpdatedAt:     createdAt,
				LastAttemptAt: frozenTime,
			},
			h: worker.JobHandler{
				Handle: func(ctx context.Context, job worker.JobSpec) error {
					return nil
				},
				JobOpts: worker.JobOptions{
					MaxAttempts: 3,
					Timeout:     time.Second,
				},
			},
			expected: worker.Job{
				Status:        worker.StatusDone,
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				LastError:     "",
			},
		},
		{
			name: "SuccessfulSecondAttempt",
			job: worker.Job{
				UpdatedAt:     createdAt,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				LastError:     "attempt 1 failed with some retryable error",
			},
			h: worker.JobHandler{
				Handle: func(ctx context.Context, job worker.JobSpec) error {
					return nil
				},
				JobOpts: worker.JobOptions{
					MaxAttempts: 3,
					Timeout:     time.Second,
				},
			},
			expected: worker.Job{
				Status:        worker.StatusDone,
				UpdatedAt:     frozenTime,
				AttemptsDone:  2,
				LastAttemptAt: frozenTime,
				LastError:     "attempt 1 failed with some retryable error",
			},
		},
		{
			name: "FailedLastAttempt",
			job: worker.Job{
				UpdatedAt:     createdAt,
				AttemptsDone:  2,
				LastAttemptAt: frozenTime,
				LastError:     "attempt 1 failed with some retryable error",
			},
			h: worker.JobHandler{
				Handle: func(ctx context.Context, job worker.JobSpec) error {
					return &worker.RetryableError{
						Cause: errors.New("some retryable error occurred"),
					}
				},
				JobOpts: worker.JobOptions{
					MaxAttempts: 3,
					Timeout:     time.Second,
				},
			},
			expected: worker.Job{
				Status:        worker.StatusDead,
				UpdatedAt:     frozenTime,
				AttemptsDone:  3,
				LastAttemptAt: frozenTime,
				LastError:     "retryable-error: some retryable error occurred",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			tc.job.Attempt(tc.ctx, frozenTime, tc.h)
			assert.Equal(t, tc.expected, tc.job)
		})
	}
}
