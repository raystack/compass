package worker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/goto/salt/log"
)

var (
	ErrTypeExists  = errors.New("handler for given job type exists")
	ErrUnknownType = errors.New("job type is invalid")
	ErrJobExists   = errors.New("job with id exists")
	ErrNoJob       = errors.New("no job found")
)

// Worker provides asynchronous job processing using a job processor.
type Worker struct {
	workers           int
	pollInterval      time.Duration
	activePollPercent float64

	processor JobProcessor
	logger    log.Logger

	mu       sync.RWMutex
	handlers map[string]JobHandler
}

type Option func(w *Worker) error

// New returns an instance of Worker initialized with defaults. By default, the
// Worker uses a noop logger with run config of 1 worker, 1s poll interval and 0
// jitter.
func New(processor JobProcessor, opts ...Option) (*Worker, error) {
	w := &Worker{
		processor: processor,
		handlers:  make(map[string]JobHandler),
	}
	for _, opt := range withDefaults(opts) {
		if err := opt(w); err != nil {
			return nil, fmt.Errorf("new worker: %w", err)
		}
	}

	return w, nil
}

// Register registers a job type and the handler that should be invoked for
// processing it.
// Returns ErrTypeExists if the type is already registered.
func (w *Worker) Register(typ string, h JobHandler) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.handlers[typ]; exists {
		return fmt.Errorf("register handler: %w: type '%s'", ErrTypeExists, typ)
	}
	if err := h.Sanitize(); err != nil {
		return fmt.Errorf("register handler: %w: type '%s'", err, typ)
	}

	w.handlers[typ] = h
	return nil
}

// Enqueue enqueues all jobs for processing.
func (w *Worker) Enqueue(ctx context.Context, jobs ...JobSpec) error {
	execs := make([]Job, 0, len(jobs))
	for _, j := range jobs {
		je, err := NewJob(j)
		if err != nil {
			return fmt.Errorf("worker enqueue: %w", err)
		}

		execs = append(execs, je)
	}

	return w.processor.Enqueue(ctx, execs...)
}

// Run starts the worker threads that dequeue and process ready jobs. Run blocks
// until all workers exit or context is canceled. Context cancellation will do
// graceful shutdown of the worker threads.
func (w *Worker) Run(baseCtx context.Context) error {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	activePollWorkers := (int)(math.Ceil((float64)(w.workers) * w.activePollPercent / 100))

	var wg sync.WaitGroup
	wg.Add(w.workers)
	for i := 0; i < w.workers; i++ {
		go func(id int) {
			defer wg.Done()

			w.runWorker(ctx, id < activePollWorkers)
			w.logger.Info("worker exited", "worker_id", id)
		}(i)
	}
	wg.Wait()

	w.logger.Info("all workers-threads exited")
	return cleanupCtxErr(ctx.Err())
}

func (w *Worker) runWorker(ctx context.Context, activePoll bool) {
	timer := time.NewTimer(w.pollInterval)
	defer timer.Stop()

	var backoff BackoffStrategy = ConstBackoff{Delay: w.pollInterval}
	if !activePoll {
		backoff = &ExponentialBackoff{
			Multiplier:   1.6,
			InitialDelay: w.pollInterval,
			MaxDelay:     5 * time.Second,
			Jitter:       0.5,
		}
	}

	pollAttempt := 1
	for {
		select {
		case <-ctx.Done():
			return

		case <-timer.C:
			types := w.getTypes()
			if len(types) == 0 {
				w.logger.Warn("no job-handler registered, skipping processing")
				continue
			}

			w.logger.Debug("looking for a job", "types", types, "active_poll", activePoll)
			switch err := w.processor.Process(ctx, types, w.processJob); {
			case err != nil && errors.Is(err, ErrNoJob):
				pollAttempt++

			case err != nil:
				w.logger.Error("process job failed", "err", err)
				pollAttempt = 1

			default:
				pollAttempt = 1
			}
			timer.Reset(backoff.Backoff(pollAttempt))
		}
	}
}

func (w *Worker) processJob(ctx context.Context, job Job) Job {
	const invalidTypeBackoff = 5 * time.Minute

	start := time.Now()
	w.logger.Info("got a job for processing",
		"job_id", job.ID,
		"job_type", job.Type,
	)

	h, ok := w.jobHandler(job.Type)
	if !ok {
		// Note: This should never happen since Process() has `Types` filter.
		//       It is only kept as a safety net to prevent nil-dereferences.
		job.LastError = ErrUnknownType.Error()
		job.RunAt = time.Now().Add(invalidTypeBackoff)
		return job
	}

	job.Attempt(ctx, time.Now(), h)

	w.logger.Info("job attempted",
		"job_id", job.ID,
		"attempts_done", job.AttemptsDone,
		"job_status", job.Status,
		"last_error", job.LastError,
		"time_ms", time.Since(start).Milliseconds(),
	)

	return job
}

func (w *Worker) jobHandler(typ string) (JobHandler, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	h, ok := w.handlers[typ]
	return h, ok
}

func (w *Worker) getTypes() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var types []string
	for typ := range w.handlers {
		types = append(types, typ)
	}
	return types
}

func cleanupCtxErr(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}
