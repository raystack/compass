package workermanager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/goto/compass/pkg/worker"
	"github.com/goto/compass/pkg/worker/pgq"
	"github.com/goto/compass/pkg/worker/workermw"
	"github.com/goto/salt/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Manager struct {
	processor      *pgq.Processor
	initDone       atomic.Bool
	worker         Worker
	jobManagerPort int
	discoveryRepo  DiscoveryRepository
	logger         log.Logger
}

//go:generate mockery --name=Worker -r --case underscore --with-expecter --structname Worker --filename worker_mock.go --output=./mocks

type Worker interface {
	Register(typ string, h worker.JobHandler) error
	Run(ctx context.Context) error
	Enqueue(ctx context.Context, jobs ...worker.JobSpec) error
}

type Config struct {
	Enabled           bool          `mapstructure:"enabled"`
	WorkerCount       int           `mapstructure:"worker_count" default:"3"`
	PollInterval      time.Duration `mapstructure:"poll_interval" default:"500ms"`
	ActivePollPercent float64       `mapstructure:"active_poll_percent" default:"20"`
	PGQ               pgq.Config    `mapstructure:"pgq"`
	JobManagerPort    int           `mapstructure:"job_manager_port"`
}

type Deps struct {
	Config        Config
	DiscoveryRepo DiscoveryRepository
	Logger        log.Logger
}

func New(ctx context.Context, deps Deps) (*Manager, error) {
	cfg := deps.Config
	processor, err := pgq.NewProcessor(ctx, cfg.PGQ)
	if err != nil {
		return nil, fmt.Errorf("new worker manager: %w", err)
	}

	w, err := worker.New(
		workermw.WithJobProcessorInstrumentation()(processor),
		worker.WithRunConfig(cfg.WorkerCount, cfg.PollInterval),
		worker.WithActivePollPercent(cfg.ActivePollPercent),
		worker.WithLogger(deps.Logger),
	)
	if err != nil {
		return nil, fmt.Errorf("new worker manager: %w", err)
	}

	return &Manager{
		processor:      processor,
		worker:         w,
		jobManagerPort: cfg.JobManagerPort,
		discoveryRepo:  deps.DiscoveryRepo,
		logger:         deps.Logger,
	}, nil
}

func NewWithWorker(w Worker, deps Deps) *Manager {
	return &Manager{
		worker:        w,
		discoveryRepo: deps.DiscoveryRepo,
	}
}

func (m *Manager) Run(ctx context.Context) error {
	if err := m.init(); err != nil {
		return fmt.Errorf("run async worker: init: %w", err)
	}

	go func() {
		srv := http.Server{
			Addr:           fmt.Sprintf(":%d", m.jobManagerPort),
			Handler:        worker.DeadJobManagementHandler(m.processor),
			ReadTimeout:    3 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.logger.Error("Worker job manager - listen and serve", "err", err)
		}
	}()

	return m.worker.Run(ctx)
}

func (m *Manager) init() error {
	if m.initDone.Load() {
		return nil
	}
	m.initDone.Store(true)

	jobHandlers := map[string]worker.JobHandler{
		jobIndexAsset:  m.indexAssetHandler(),
		jobDeleteAsset: m.deleteAssetHandler(),
	}
	for typ, h := range jobHandlers {
		if err := m.worker.Register(typ, h); err != nil {
			return err
		}
	}

	return m.registerStatsCallback(keys(jobHandlers))
}

func (m *Manager) Close() error {
	return m.processor.Close()
}

func (m *Manager) registerStatsCallback(jobTypes []string) error {
	const attrJobType = attribute.Key("job.type")

	meter := otel.Meter("github.com/goto/compass/internal/workermanager")
	activeJobs, err := meter.Int64ObservableGauge("compass.worker.active_jobs")
	handleOtelErr(err)

	deadJobs, err := meter.Int64ObservableGauge("compass.worker.dead_jobs")
	handleOtelErr(err)

	_, err = meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			stats, err := m.processor.Stats(ctx)
			if err != nil {
				return err
			}

			seen := make(map[string]struct{}, len(jobTypes))
			for _, st := range stats {
				seen[st.Type] = struct{}{}
				attr := metric.WithAttributes(attrJobType.String(st.Type))
				o.ObserveInt64(activeJobs, (int64)(st.Active), attr)
				o.ObserveInt64(deadJobs, (int64)(st.Dead), attr)
			}

			for _, typ := range jobTypes {
				if _, ok := seen[typ]; ok {
					continue
				}

				attr := metric.WithAttributes(attrJobType.String(typ))
				o.ObserveInt64(activeJobs, 0, attr)
				o.ObserveInt64(deadJobs, 0, attr)
			}

			return nil
		},
		activeJobs,
		deadJobs,
	)

	return err
}

func keys(handlers map[string]worker.JobHandler) []string {
	types := make([]string, 0, len(handlers))
	for typ := range handlers {
		types = append(types, typ)
	}
	return types
}

func handleOtelErr(err error) {
	if err != nil {
		otel.Handle(err)
	}
}
