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
	"github.com/goto/salt/log"
)

type Manager struct {
	processor      *pgq.Processor
	registered     atomic.Bool
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
	Enabled        bool          `mapstructure:"enabled"`
	WorkerCount    int           `mapstructure:"worker_count" default:"3"`
	PollInterval   time.Duration `mapstructure:"poll_interval" default:"500ms"`
	PGQ            pgq.Config    `mapstructure:"pgq"`
	JobManagerPort int           `mapstructure:"job_manager_port"`
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
		processor,
		worker.WithRunConfig(cfg.WorkerCount, cfg.PollInterval),
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
	if err := m.register(); err != nil {
		return fmt.Errorf("run async worker: %w", err)
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

func (m *Manager) register() error {
	if m.registered.Load() {
		return nil
	}

	for typ, h := range map[string]worker.JobHandler{
		jobIndexAsset:  m.indexAssetHandler(),
		jobDeleteAsset: m.deleteAssetHandler(),
	} {
		if err := m.worker.Register(typ, h); err != nil {
			return err
		}
	}

	m.registered.Store(true)

	return nil
}

func (m *Manager) Close() error {
	return m.processor.Close()
}
