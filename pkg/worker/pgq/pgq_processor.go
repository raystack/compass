package pgq

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/goto/compass/pkg/worker"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib" // register driver
)

const (
	pgDriverName  = "pgx"
	jobsTable     = "jobs_queue"
	deadJobsTable = "dead_jobs"
)

// Processor implements a JobProcessor backed by PostgreSQL.
type Processor struct {
	db *sql.DB
}

// NewProcessor returns a JobProcessor implementation backed by the PostgreSQL
// instance identified by the provided config.
func NewProcessor(ctx context.Context, cfg Config) (*Processor, error) {
	db, err := sql.Open(pgDriverName, cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("new pgq processor: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("new pgq processor: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	if cfg.MaxIdleConns != 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if maxLifetime := cfg.ConnMaxLifetimeWithJitter(); maxLifetime != 0 {
		db.SetConnMaxLifetime(maxLifetime)
	}
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return &Processor{db: db}, nil
}

func (p *Processor) Enqueue(ctx context.Context, jobs ...worker.Job) error {
	insert := sq.Insert(jobsTable).Columns(
		"id", "type", "run_at", "payload", "created_at", "updated_at",
	)
	for _, j := range jobs {
		insert = insert.Values(
			j.ID.String(), j.Type, j.RunAt.UTC(), j.Payload, j.CreatedAt.UTC(), j.UpdatedAt.UTC(),
		)
	}

	_, err := insert.RunWith(p.db).
		PlaceholderFormat(sq.Dollar).
		ExecContext(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return fmt.Errorf("enqueue jobs: %w: %s", worker.ErrJobExists, err.Error())
			}
		}
		return fmt.Errorf("enqueue jobs: %w", err)
	}

	return nil
}

func (p *Processor) Process(ctx context.Context, types []string, fn worker.JobExecutorFunc) error {
	err := p.withTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		job, err := p.pickupJob(ctx, tx, types)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("pickup job: %w", err)
		}

		resultJob := fn(ctx, job)
		switch resultJob.Status {
		case worker.StatusDone:
			if err := p.clearJob(ctx, tx, resultJob); err != nil {
				return err
			}

		case worker.StatusDead:
			if err := p.markJobDead(ctx, tx, resultJob); err != nil {
				return err
			}

		default:
			if err := p.setupRetry(ctx, tx, resultJob); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("pgq process: %w", err)
	}
	return nil
}

func (p *Processor) Close() error { return p.db.Close() }
