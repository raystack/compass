package worker

import (
	"time"

	"github.com/goto/salt/log"
)

func WithJobHandler(typ string, h JobHandler) Option {
	return func(w *Worker) error {
		return w.Register(typ, h)
	}
}

func WithLogger(l log.Logger) Option {
	return func(w *Worker) error {
		if l == nil {
			l = log.NewNoop()
		}
		w.logger = l
		return nil
	}
}

func WithRunConfig(workers int, pollInterval time.Duration) Option {
	return func(w *Worker) error {
		if workers == 0 {
			workers = 1
		}

		const minPollInterval = 100 * time.Millisecond
		if pollInterval < minPollInterval {
			pollInterval = minPollInterval
		}

		w.workers = workers
		w.pollInterval = pollInterval
		return nil
	}
}

func WithActivePollPercent(pct float64) Option {
	return func(w *Worker) error {
		w.activePollPercent = pct
		return nil
	}
}

func withDefaults(opts []Option) []Option {
	return append([]Option{
		WithLogger(nil),
		WithRunConfig(1, 1*time.Second),
		WithActivePollPercent(20),
	}, opts...)
}
