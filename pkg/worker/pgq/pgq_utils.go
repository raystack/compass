package pgq

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/goto/compass/pkg/worker"
	"github.com/oklog/ulid/v2"
)

func (p *Processor) withTx(ctx context.Context, fn func(context.Context, *sql.Tx) error) (err error) {
	var tx *sql.Tx
	tx, err = p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("run with tx: %w", err)
	}

	return tx.Commit()
}

func (*Processor) pickupJob(ctx context.Context, r sq.BaseRunner, types []string) (worker.Job, error) {
	query := sq.Select().
		From(jobsTable).
		Columns(
			"id", "type", "run_at", "payload", "created_at",
			"updated_at", "attempts_done", "last_attempt_at", "last_error",
		).
		Where(sq.Eq{"type": types}).
		Where(sq.Expr("run_at <= current_timestamp")).
		OrderBy("id ASC").
		Limit(1).
		Suffix("FOR UPDATE SKIP LOCKED")

	var (
		job           worker.Job
		id            string
		lastErr       sql.NullString
		lastAttemptAt sql.NullTime
	)
	err := query.PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRowContext(ctx).
		Scan(
			&id, &job.Type, &job.RunAt, &job.Payload, &job.CreatedAt,
			&job.UpdatedAt, &job.AttemptsDone, &lastAttemptAt, &lastErr,
		)
	if err != nil {
		return worker.Job{}, fmt.Errorf("scan row: %w", err)
	}

	uid, err := ulid.ParseStrict(id)
	if err != nil {
		return worker.Job{}, fmt.Errorf("scan row: parse ULID: %w", err)
	}

	job.ID = uid
	job.LastAttemptAt = lastAttemptAt.Time
	job.LastError = lastErr.String

	return job, nil
}

func (*Processor) clearJob(ctx context.Context, r sq.BaseRunner, job worker.Job) error {
	query := sq.Delete(jobsTable).
		Where(sq.Eq{"id": job.ID.String()})

	res, err := query.PlaceholderFormat(sq.Dollar).
		RunWith(r).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("clear job: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("clear job: check rows affected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("clear job: rows affected: %d", rowsAffected)
	}

	return nil
}

func (p *Processor) markJobDead(ctx context.Context, r sq.BaseRunner, job worker.Job) error {
	insert := sq.Insert(deadJobsTable).
		Columns(
			"id", "type", "payload", "created_at",
			"updated_at", "attempts_done", "last_attempt_at", "last_error",
		).
		Values(
			job.ID.String(), job.Type, job.Payload, job.CreatedAt.UTC(),
			job.UpdatedAt.UTC(), job.AttemptsDone, job.LastAttemptAt.UTC(), job.LastError,
		)

	_, err := insert.PlaceholderFormat(sq.Dollar).
		RunWith(r).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("mark job as dead: %w", err)
	}

	if err := p.clearJob(ctx, r, job); err != nil {
		return fmt.Errorf("mark job as dead: %w", err)
	}

	return nil
}

func (*Processor) setupRetry(ctx context.Context, r sq.BaseRunner, job worker.Job) error {
	update := sq.Update(jobsTable).
		Where(sq.Eq{"id": job.ID.String()}).
		Set("run_at", job.RunAt.UTC()).
		Set("updated_at", job.UpdatedAt.UTC()).
		Set("attempts_done", sq.Expr("attempts_done + 1")).
		Set("last_error", job.LastError).
		Set("last_attempt_at", job.LastAttemptAt.UTC())

	res, err := update.PlaceholderFormat(sq.Dollar).
		RunWith(r).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("setup job retry: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("setup job retry: check rows affected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("setup job retry: rows affected: %d", rowsAffected)
	}
	return nil
}
