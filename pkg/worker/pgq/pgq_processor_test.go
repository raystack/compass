package pgq_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/goto/compass/internal/testutils"
	"github.com/goto/compass/pkg/worker"
	"github.com/goto/compass/pkg/worker/pgq"
	"github.com/goto/salt/log"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/suite"
)

type ProcessorTestSuite struct {
	suite.Suite
	ctx    context.Context
	db     *sql.DB
	pgPort int
}

func TestProcessor(t *testing.T) {
	suite.Run(t, &ProcessorTestSuite{})
}

func (s *ProcessorTestSuite) SetupSuite() {
	logger := log.NewLogrus()
	port, err := testutils.RunTestPG(s.T(), logger)
	s.Require().NoError(err)
	s.pgPort = port

	db, err := sql.Open("pgx", s.testDBConfig().ConnectionString())
	s.Require().NoError(err)

	s.T().Cleanup(func() {
		s.Require().NoError(db.Close())
	})

	s.ctx = context.Background()
	s.db = db
}

func (s *ProcessorTestSuite) TestNewProcessor() {
	s.Run("InvalidConfig", func() {
		cfg := s.testDBConfig()
		cfg.Port++

		p, err := pgq.NewProcessor(s.ctx, cfg)
		s.ErrorContains(err, "new pgq processor: failed to connect")
		s.Nil(p)
	})

	s.Run("Success", func() {
		p, err := pgq.NewProcessor(s.ctx, s.testDBConfig())
		s.NoError(err)
		s.NotNil(p)
		s.NoError(p.Close())
	})
}

func (s *ProcessorTestSuite) TestEnqueue() {
	jobSpecs := []worker.JobSpec{
		{Type: "test", Payload: []byte("job1")},
		{Type: "test", Payload: []byte("job2")},
		{Type: "test", Payload: []byte("job3")},
	}
	var jobs []worker.Job
	for _, js := range jobSpecs {
		job, err := worker.NewJob(js)
		s.Require().NoError(err)

		job.RunAt = job.RunAt.UTC().Truncate(time.Second)
		job.CreatedAt = job.CreatedAt.UTC().Truncate(time.Second)
		job.UpdatedAt = job.UpdatedAt.UTC().Truncate(time.Second)
		jobs = append(jobs, job)
	}

	p, err := pgq.NewProcessor(s.ctx, s.testDBConfig())
	s.NoError(err)

	defer p.Close()
	cases := []struct {
		name        string
		jobs        []worker.Job
		expected    []worker.Job
		expectedErr error
	}{
		{
			name:     "SingleJob",
			jobs:     []worker.Job{jobs[0]},
			expected: []worker.Job{jobs[0]},
		},
		{
			name:     "MultipleJobs",
			jobs:     jobs,
			expected: jobs,
		},
		{
			name:        "DuplicateJobs",
			jobs:        []worker.Job{jobs[0], jobs[1], jobs[0]},
			expectedErr: worker.ErrJobExists,
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			err := testutils.RunMigrations(s.T(), s.db)
			s.Require().NoError(err)

			if err := p.Enqueue(s.ctx, tc.jobs...); tc.expectedErr != nil {
				s.ErrorIs(err, tc.expectedErr)
			} else {
				s.NoError(err)
			}

			query := "SELECT id, type, run_at, payload, created_at, " +
				"updated_at, attempts_done, last_attempt_at, last_error " +
				"FROM jobs_queue"
			rows, err := s.db.Query(query)
			s.Require().NoError(err)
			defer rows.Close()

			var actual []worker.Job
			for rows.Next() {
				actual = append(actual, s.scanJob(rows))
			}
			s.Require().NoError(rows.Err())

			s.Equal(tc.expected, actual)
		})
	}
}

func (s *ProcessorTestSuite) TestProcess() {
	frozenTime := time.Unix(1654082526, 0).UTC()

	job, err := worker.NewJob(worker.JobSpec{Type: "test", Payload: []byte("payload")})
	s.Require().NoError(err)

	job.RunAt = frozenTime
	job.CreatedAt = frozenTime
	job.UpdatedAt = frozenTime

	p, err := pgq.NewProcessor(s.ctx, s.testDBConfig())
	s.NoError(err)

	s.Run("NoJobs", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		err = p.Process(s.ctx, []string{"test"}, func(ctx context.Context, job worker.Job) worker.Job {
			s.Fail("unexpected job invocation")
			return job
		})
		s.NoError(err)
	})

	s.Run("ProcessOnlyGivenTypes", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		err = p.Enqueue(s.ctx, job)
		s.NoError(err)

		err = p.Process(s.ctx, []string{"new-type"}, func(ctx context.Context, job worker.Job) worker.Job {
			s.Fail("unexpected job invocation")
			return job
		})
		s.NoError(err)
	})

	s.Run("ProcessOnlyReadyJobs", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		job := job
		job.RunAt = time.Now().AddDate(0, 0, 1)
		err = p.Enqueue(s.ctx, job)
		s.NoError(err)

		err = p.Process(s.ctx, []string{"type"}, func(ctx context.Context, job worker.Job) worker.Job {
			s.Fail("unexpected job invocation")
			return job
		})
		s.NoError(err)
	})

	s.Run("JobProcessedSuccessfully", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		err = p.Enqueue(s.ctx, job)
		s.NoError(err)

		err = p.Process(s.ctx, []string{"test"}, func(ctx context.Context, job worker.Job) worker.Job {
			job.AttemptsDone++
			job.Status = worker.StatusDone
			job.LastAttemptAt = frozenTime.Add(time.Second * 5)
			return job
		})
		s.NoError(err)

		query := "SELECT count(*) FROM jobs_queue WHERE id = $1"
		var cnt int
		err = s.db.QueryRow(query, job.ID.String()).Scan(&cnt)
		s.Require().NoError(err)

		s.Zero(cnt)

		query = "SELECT count(*) FROM dead_jobs WHERE id = $1"
		err = s.db.QueryRow(query, job.ID.String()).Scan(&cnt)
		s.Require().NoError(err)

		s.Zero(cnt)
	})

	s.Run("DeadJob", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		err = p.Enqueue(s.ctx, job)
		s.NoError(err)

		var deadJob worker.Job
		err = p.Process(s.ctx, []string{"test"}, func(ctx context.Context, job worker.Job) worker.Job {
			job.AttemptsDone++
			job.Status = worker.StatusDead
			job.LastAttemptAt = frozenTime.Add(time.Second * 5)
			job.LastError = "murdered"
			deadJob = job
			return job
		})
		s.NoError(err)

		query := "SELECT count(*) FROM jobs_queue WHERE id = $1"
		var cnt int
		err = s.db.QueryRow(query, job.ID.String()).Scan(&cnt)
		s.Require().NoError(err)

		s.Zero(cnt)

		query = "SELECT id, type, payload, created_at, " +
			"updated_at, attempts_done, last_attempt_at, last_error " +
			"FROM dead_jobs WHERE id = $1"

		var (
			actual        worker.Job
			id            string
			lastErr       sql.NullString
			lastAttemptAt sql.NullTime
		)
		err = s.db.QueryRow(query, job.ID.String()).Scan(
			&id, &actual.Type, &actual.Payload, &actual.CreatedAt,
			&actual.UpdatedAt, &actual.AttemptsDone, &lastAttemptAt, &lastErr,
		)
		s.Require().NoError(err)

		uid, err := ulid.ParseStrict(id)
		s.Require().NoError(err)

		actual.ID = uid
		actual.LastAttemptAt = lastAttemptAt.Time
		actual.LastError = lastErr.String

		deadJob.RunAt = time.Time{}
		deadJob.Status = ""
		s.Equal(deadJob, actual)
	})

	s.Run("JobRetry", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		err = p.Enqueue(s.ctx, job)
		s.NoError(err)

		var retryJob worker.Job
		err = p.Process(s.ctx, []string{"test"}, func(ctx context.Context, job worker.Job) worker.Job {
			job.AttemptsDone++
			job.LastAttemptAt = frozenTime.Add(time.Second * 5)
			job.RunAt = frozenTime.Add(time.Second * 10)
			job.LastError = "attempted murder"
			retryJob = job
			return job
		})
		s.NoError(err)

		query := "SELECT id, type, run_at, payload, created_at, " +
			"updated_at, attempts_done, last_attempt_at, last_error " +
			"FROM jobs_queue WHERE id = $1"
		row := s.db.QueryRow(query, job.ID.String())
		actual := s.scanJob(row)
		s.Equal(retryJob, actual)

		query = "SELECT count(*) FROM dead_jobs WHERE id = $1"
		var cnt int
		err = s.db.QueryRow(query, job.ID.String()).Scan(&cnt)
		s.Require().NoError(err)

		s.Zero(cnt)
	})

	s.Run("JobLocking", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		err = p.Enqueue(s.ctx, job)
		s.NoError(err)

		var jobInvoked bool
		err = p.Process(s.ctx, []string{"test"}, func(ctx context.Context, job worker.Job) worker.Job {
			jobInvoked = true
			err := p.Process(s.ctx, []string{"test"}, func(ctx context.Context, job worker.Job) worker.Job {
				s.Fail("unexpected job invocation")
				return job
			})
			s.NoError(err)
			return job
		})
		s.NoError(err)
		s.True(jobInvoked)
	})

	s.Run("Rollback", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		err = p.Enqueue(s.ctx, job)
		s.NoError(err)

		err = p.Process(s.ctx, []string{"test"}, func(ctx context.Context, job worker.Job) worker.Job {
			panic("die")
		})
		s.EqualError(err, "pgq process: panic: die")

		query := "SELECT count(*) FROM jobs_queue WHERE id = $1"
		var cnt int
		err = s.db.QueryRow(query, job.ID.String()).Scan(&cnt)
		s.Require().NoError(err)
		s.Equal(1, cnt)
	})

	s.Run("ClearJobFailure", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		err = p.Enqueue(s.ctx, job)
		s.NoError(err)

		err = p.Process(s.ctx, []string{"test"}, func(ctx context.Context, job worker.Job) worker.Job {
			job.ID = ulid.Make()
			job.AttemptsDone++
			job.Status = worker.StatusDone
			job.LastAttemptAt = frozenTime.Add(time.Second * 5)
			return job
		})
		s.EqualError(err, "pgq process: run with tx: clear job: rows affected: 0")
	})

	s.Run("DuplicateDeadJob", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		fn := func(ctx context.Context, job worker.Job) worker.Job {
			job.AttemptsDone++
			job.Status = worker.StatusDead
			job.LastAttemptAt = frozenTime.Add(time.Second * 5)
			job.LastError = "murdered"
			return job
		}
		err = p.Enqueue(s.ctx, job)
		s.NoError(err)
		err = p.Process(s.ctx, []string{"test"}, fn)
		s.NoError(err)

		err = p.Enqueue(s.ctx, job)
		s.NoError(err)
		err = p.Process(s.ctx, []string{"test"}, fn)
		s.ErrorContains(err, `mark job as dead: ERROR: duplicate key value violates unique constraint "dead_jobs_pkey"`)
	})

	s.Run("UpdateJobFailure", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		err = p.Enqueue(s.ctx, job)
		s.NoError(err)

		err = p.Process(s.ctx, []string{"test"}, func(ctx context.Context, job worker.Job) worker.Job {
			job.ID = ulid.Make()
			job.AttemptsDone++
			job.LastAttemptAt = frozenTime.Add(time.Second * 5)
			job.RunAt = frozenTime.Add(time.Second * 10)
			job.LastError = "attempted murder"
			return job
		})
		s.EqualError(err, "pgq process: run with tx: setup job retry: rows affected: 0")
	})
}

func (s *ProcessorTestSuite) scanJob(row interface{ Scan(...interface{}) error }) worker.Job {
	s.T().Helper()

	var (
		job           worker.Job
		id            string
		lastErr       sql.NullString
		lastAttemptAt sql.NullTime
	)
	err := row.Scan(
		&id, &job.Type, &job.RunAt, &job.Payload, &job.CreatedAt,
		&job.UpdatedAt, &job.AttemptsDone, &lastAttemptAt, &lastErr,
	)
	s.Require().NoError(err)

	uid, err := ulid.ParseStrict(id)
	s.Require().NoError(err)

	job.ID = uid
	job.LastAttemptAt = lastAttemptAt.Time
	job.LastError = lastErr.String

	return job
}

func (s *ProcessorTestSuite) testDBConfig() pgq.Config {
	s.T().Helper()

	return pgq.Config{
		Host:     testutils.PGHost,
		Port:     s.pgPort,
		Name:     testutils.PGName,
		Username: testutils.PGUsername,
		Password: testutils.PGPassword,
	}
}
