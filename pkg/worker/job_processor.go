package worker

import "context"

//go:generate mockery --name=JobProcessor -r --case underscore --with-expecter --structname JobProcessor --filename job_processor_mock.go --output=./mocks

// JobProcessor represents a special job store or queue that holds jobs and
// processes them via Process() only after the jobs are ready.
type JobProcessor interface {
	// Enqueue all jobs. Enqueue must ensure all-or-nothing behavior.
	// Jobs with zero-value or historical value for ReadyAt must be executed
	// immediately.
	Enqueue(ctx context.Context, jobs ...Job) error

	// Process dequeues one job from the data store and invokes `fn`. The job
	// should be 'locked' until `fn` returns. Refer JobExecutorFunc.
	// Process is also responsible for clearing the job or marking the job as
	// dead or setting up the retry for the job depending on the job result.
	Process(ctx context.Context, types []string, fn JobExecutorFunc) error

	// Stats returns the job statistics by job type. It includes the number of
	// active and dead jobs.
	Stats(ctx context.Context) ([]JobTypeStats, error)
}

// JobTypeStats is the statistics for the job type with number of active and
// dead jobs.
type JobTypeStats struct {
	Type   string `json:"type"`
	Active int    `json:"active"`
	Dead   int    `json:"dead"`
}

// JobExecutorFunc is invoked by JobProcessor for ready jobs. It is responsible
// for handling a ready job and returning the updated job execution result after
// the attempt.
type JobExecutorFunc func(context.Context, Job) Job
