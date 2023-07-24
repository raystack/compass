package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/goto/compass/pkg/worker"
	"github.com/goto/compass/pkg/worker/mocks"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Parallel()

	var p mocks.JobProcessor
	h := worker.JobHandler{Handle: func(ctx context.Context, spec worker.JobSpec) error {
		return nil
	}}

	t.Run("DuplicateType", func(t *testing.T) {
		w, err := worker.New(&p,
			worker.WithJobHandler("test", h),
			worker.WithJobHandler("test", h),
		)
		assert.EqualError(t, err, "new worker: register handler: handler for given job type exists: type 'test'")
		assert.Nil(t, w)
	})

	t.Run("MissingJobHandler", func(t *testing.T) {
		w, err := worker.New(&p,
			worker.WithJobHandler("test", worker.JobHandler{}),
		)
		assert.EqualError(t, err, "new worker: register handler: sanitize job handler: "+
			"job handler is not valid: handle function must be set: type 'test'")
		assert.Nil(t, w)
	})

	t.Run("Success", func(t *testing.T) {
		w, err := worker.New(&p,
			worker.WithJobHandler("test", h),
			worker.WithRunConfig(0, 0),
		)
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})
}

func TestWorker_Enqueue(t *testing.T) {
	t.Parallel()

	ts := time.Unix(1654082526, 0)
	table := []struct {
		name        string
		queue       func(t *testing.T) worker.JobProcessor
		opts        []worker.Option
		jobs        []worker.JobSpec
		expectedErr error
	}{
		{
			name: "WithoutType",
			queue: func(t *testing.T) worker.JobProcessor {
				t.Helper()
				return mocks.NewJobProcessor(t)
			},
			jobs:        []worker.JobSpec{{}},
			expectedErr: worker.ErrInvalidJob,
		},
		{
			name: "Success",
			queue: func(t *testing.T) worker.JobProcessor {
				t.Helper()

				now := time.Now()
				q := mocks.NewJobProcessor(t)
				q.EXPECT().
					Enqueue(
						mock.Anything,
						mock.AnythingOfType("worker.Job"),
						mock.AnythingOfType("worker.Job"),
						mock.AnythingOfType("worker.Job"),
					).
					Run(func(ctx context.Context, jobs ...worker.Job) {
						require.Len(t, jobs, 3)
						assert.NotEmpty(t, jobs[0].ID)
						assert.Equal(t, "test-1", jobs[0].Type)
						assert.WithinDuration(t, jobs[0].RunAt, now, 5*time.Second)
						assert.WithinDuration(t, jobs[0].CreatedAt, now, 5*time.Second)
						assert.WithinDuration(t, jobs[0].UpdatedAt, now, 5*time.Second)
						assert.NotEmpty(t, jobs[1].ID)
						assert.Equal(t, "test-2", jobs[1].Type)
						assert.WithinDuration(t, jobs[1].RunAt, now, 5*time.Second)
						assert.WithinDuration(t, jobs[1].CreatedAt, now, 5*time.Second)
						assert.WithinDuration(t, jobs[1].UpdatedAt, now, 5*time.Second)
						assert.NotEmpty(t, jobs[2].ID)
						assert.Equal(t, "test-3", jobs[2].Type)
						assert.Equal(t, jobs[2].RunAt, ts)
						assert.WithinDuration(t, jobs[2].CreatedAt, now, 5*time.Second)
						assert.WithinDuration(t, jobs[2].UpdatedAt, now, 5*time.Second)
					}).
					Return(nil).
					Once()
				return q
			},
			opts: []worker.Option{
				worker.WithJobHandler("test", worker.JobHandler{Handle: func(ctx context.Context, job worker.JobSpec) error {
					return nil
				}}),
			},
			jobs: []worker.JobSpec{
				{Type: "test-1"},
				{Type: "test-2"},
				{Type: "test-3", RunAt: ts},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			w, err := worker.New(tc.queue(t), tc.opts...)
			require.NoError(t, err)
			require.NotNil(t, w)

			expected := w.Enqueue(context.Background(), tc.jobs...)
			if tc.expectedErr != nil {
				assert.Error(t, expected)
				assert.True(t, errors.Is(expected, tc.expectedErr))
			} else {
				assert.NoError(t, expected)
			}
		})
	}
}

func TestWorker_Run(t *testing.T) {
	t.Parallel()

	opts := []worker.Option{
		worker.WithJobHandler("test", worker.JobHandler{Handle: func(ctx context.Context, job worker.JobSpec) error {
			return nil
		}}),
		worker.WithRunConfig(1, 10*time.Millisecond),
	}

	t.Run("ContextCancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // immediately cancel the context.

		p := mocks.NewJobProcessor(t)

		w, err := worker.New(p, opts...)
		require.NoError(t, err)
		require.NotNil(t, w)

		expected := w.Run(ctx)
		assert.NoError(t, expected)
	})

	t.Run("ContextDeadline", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
		defer cancel()

		q := mocks.NewJobProcessor(t)

		w, err := worker.New(q, opts...)
		require.NoError(t, err)
		require.NotNil(t, w)

		expected := w.Run(ctx)
		assert.NoError(t, expected)
	})

	t.Run("ProcessReturnsUnknownType", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dequeued := 0
		sampleJob := worker.Job{
			JobSpec: worker.JobSpec{
				Type: "unknown_type",
			},
		}

		q := mocks.NewJobProcessor(t)
		q.EXPECT().
			Process(mock.Anything, []string{"test"}, mock.Anything).
			Run(func(ctx context.Context, types []string, fn worker.JobExecutorFunc) {
				resultJob := fn(ctx, sampleJob)
				assert.Equal(t, resultJob.LastError, "job type is invalid")
				assert.WithinDuration(t, time.Now().Add(5*time.Minute), resultJob.RunAt, 5*time.Second)

				dequeued++
				cancel() // cancel context to stop the worker.
			}).
			Return(nil)

		w, err := worker.New(q, opts...)
		require.NoError(t, err)
		require.NotNil(t, w)

		expected := w.Run(ctx)
		assert.NoError(t, expected)
		assert.Equal(t, 1, dequeued)
	})

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dequeued := 0
		sampleJob := worker.Job{
			ID: ulid.Make(),
			JobSpec: worker.JobSpec{
				Type: "test",
			},
		}

		q := mocks.NewJobProcessor(t)
		q.EXPECT().
			Process(mock.Anything, []string{"test"}, mock.Anything).
			Run(func(ctx context.Context, types []string, fn worker.JobExecutorFunc) {
				resultJob := fn(ctx, sampleJob)
				assert.EqualValues(t, worker.StatusDone, resultJob.Status)

				dequeued++
				cancel() // cancel context to stop the worker.
			}).
			Return(nil)

		w, err := worker.New(q, opts...)
		require.NoError(t, err)
		require.NotNil(t, w)

		expected := w.Run(ctx)
		assert.NoError(t, expected)
		assert.Equal(t, 1, dequeued)
	})
}
