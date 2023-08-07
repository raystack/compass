package pgq_test

import (
	"context"
	"database/sql"
	"os"
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

		actual, err := p.DeadJobs(s.ctx, 1, 0)
		s.Require().NoError(err)

		deadJob.RunAt = time.Time{}
		deadJob.Status = ""
		s.Equal(deadJob, actual[0])
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

func (s *ProcessorTestSuite) TestStats() {
	s.Run("WithEmptyTables", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		p, err := pgq.NewProcessor(s.ctx, s.testDBConfig())
		s.NoError(err)

		stats, err := p.Stats(s.ctx)
		s.NoError(err)
		s.Empty(stats)
	})

	s.Run("WithActiveAndDeadJobs", func() {
		err := testutils.RunMigrations(s.T(), s.db)
		s.Require().NoError(err)

		inserts, err := os.ReadFile("testdata/insert_dead_jobs.sql")
		s.Require().NoError(err)

		_, err = s.db.Exec((string)(inserts))
		s.Require().NoError(err)

		p, err := pgq.NewProcessor(s.ctx, s.testDBConfig())
		s.NoError(err)

		jobSpecs := []worker.JobSpec{
			{Type: "test", Payload: []byte("job1")},
			{Type: "test", Payload: []byte("job2")},
			{Type: "test", Payload: []byte("job3")},
		}
		var jobs []worker.Job
		for _, js := range jobSpecs {
			job, err := worker.NewJob(js)
			s.Require().NoError(err)

			jobs = append(jobs, job)
		}
		err = p.Enqueue(s.ctx, jobs...)
		s.NoError(err)

		stats, err := p.Stats(s.ctx)
		s.NoError(err)

		expected := []worker.JobTypeStats{
			{
				Type:   "index-asset",
				Active: 0,
				Dead:   12,
			},
			{
				Type:   "test",
				Active: 3,
				Dead:   0,
			},
		}
		s.Equal(expected, stats)

		err = p.Resurrect(s.ctx, []string{"01H63BPX1D3S98K5GBADE1Q322", "01H63BPVMH6BBQJH776KRJAMH0"})
		s.NoError(err)

		stats, err = p.Stats(s.ctx)
		s.NoError(err)

		expected = []worker.JobTypeStats{
			{
				Type:   "index-asset",
				Active: 2,
				Dead:   10,
			},
			{
				Type:   "test",
				Active: 3,
				Dead:   0,
			},
		}
		s.Equal(expected, stats)
	})
}

func (s *ProcessorTestSuite) TestDeadJobs() {
	err := testutils.RunMigrations(s.T(), s.db)
	s.Require().NoError(err)

	inserts, err := os.ReadFile("testdata/insert_dead_jobs.sql")
	s.Require().NoError(err)

	_, err = s.db.Exec((string)(inserts))
	s.Require().NoError(err)

	p, err := pgq.NewProcessor(s.ctx, s.testDBConfig())
	s.NoError(err)

	cases := []struct {
		name   string
		size   int
		offset int
		ids    []string
	}{
		{
			name: "Size=5;Offset=0",
			size: 5,
			ids: []string{
				"01H63BPCE1JVJHDP1WS9845MN9", "01H63BPVMH6BBQJH776KRJAMH0",
				"01H63BPX1D3S98K5GBADE1Q322", "01H63BPY8D4V5SM81KCT5KBW7V",
				"01H63BPZCB1H6K66M86RY8G85W",
			},
		},
		{
			name:   "Size=2;Offset=10",
			size:   2,
			offset: 10,
			ids:    []string{"01H63BQ5D0SZEWVX2S05K0A35C", "01H63BQ6A51PQVECDANDCXY8Y8"},
		},
		{
			name:   "Size=4;Offset=10",
			size:   4,
			offset: 10,
			ids:    []string{"01H63BQ5D0SZEWVX2S05K0A35C", "01H63BQ6A51PQVECDANDCXY8Y8"},
		},
		{
			name:   "Size=10;Offset=19",
			size:   10,
			offset: 19,
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			jobs, err := p.DeadJobs(s.ctx, tc.size, tc.offset)
			s.NoError(err)

			var ids []string
			for _, j := range jobs {
				ids = append(ids, j.ID.String())
			}
			s.Equal(tc.ids, ids)
		})
	}

	s.Run("Fields", func() {
		jobs, err := p.DeadJobs(s.ctx, 1, 0)
		s.NoError(err)

		job := jobs[0]
		s.Equal("01H63BPCE1JVJHDP1WS9845MN9", job.ID.String())
		s.Equal("index-asset", job.Type)
		s.JSONEq(
			`{"id":"6d99aab9-dc60-4e92-9c08-8b9e0ad148ea","urn":"urn:firehose:p-godata-id:job:p-godata-id-go-food-pickup-booking-log-bigquery-firehose","type":"job","service":"firehose","name":"p-godata-id-go-food-pickup-booking-log-bigquery-firehose","description":"Migrated from system?","data":{},"url":"","labels":null,"created_at":"2023-07-24T12:39:22.215737+05:30","updated_at":"2023-07-24T12:39:22.215737+05:30","version":"0.1","updated_by":{"email":"","provider":""}}`,
			(string)(job.Payload),
		)
		s.WithinDuration(
			time.Date(2023, time.July, 24, 7, 9, 22, 241788*1000, time.UTC), // 2023-07-24 07:09:22.241788
			job.CreatedAt,
			time.Millisecond,
		)
		s.WithinDuration(
			time.Date(2023, time.July, 24, 7, 9, 22, 668796*1000, time.UTC), // 2023-07-24 07:09:22.241788
			job.UpdatedAt,
			time.Millisecond,
		)
		s.Equal(3, job.AttemptsDone)
		s.WithinDuration(
			time.Date(2023, time.July, 24, 7, 9, 22, 668796*1000, time.UTC), // 2023-07-24 07:09:22.241788
			job.LastAttemptAt,
			time.Millisecond,
		)
		s.Equal("fail 1", job.LastError)
	})
}

func (s *ProcessorTestSuite) TestResurrect() {
	err := testutils.RunMigrations(s.T(), s.db)
	s.Require().NoError(err)

	inserts, err := os.ReadFile("testdata/insert_dead_jobs.sql")
	s.Require().NoError(err)

	_, err = s.db.Exec((string)(inserts))
	s.Require().NoError(err)

	p, err := pgq.NewProcessor(s.ctx, s.testDBConfig())
	s.NoError(err)

	s.Run("InvalidID", func() {
		err := p.Resurrect(s.ctx, []string{"01H59WDEB4FDS83K8HFNGVJQ7G"})
		s.NoError(err)

		query := "SELECT count(*) FROM jobs_queue WHERE id = $1"
		var cnt int
		err = s.db.QueryRow(query, "01H59WDEB4FDS83K8HFNGVJQ7G").Scan(&cnt)
		s.Require().NoError(err)
		s.Equal(0, cnt)
	})

	s.Run("ValidIDs", func() {
		resurrectionIDs := []string{
			"01H63BPCE1JVJHDP1WS9845MN9", "01H63BPVMH6BBQJH776KRJAMH0",
			"01H63BPX1D3S98K5GBADE1Q322", "01H63BPY8D4V5SM81KCT5KBW7V",
			"01H63BPZCB1H6K66M86RY8G85W", "01H63BQ0FR4TJP6431GY0FY3YY",
		}
		err := p.Resurrect(s.ctx, resurrectionIDs)
		s.NoError(err)

		jobs, err := p.DeadJobs(s.ctx, 20, 0)
		s.NoError(err)

		var ids []string
		for _, j := range jobs {
			ids = append(ids, j.ID.String())
		}
		stillDead := []string{
			"01H63BQ1HPAT44WS70EG145GKJ", "01H63BQ2KVAZEK6DXTQC4D61HP",
			"01H63BQ3KY6V76FAWAJYVKGDZK", "01H63BQ4GWT8P9BZ8NYZJKMD4W",
			"01H63BQ5D0SZEWVX2S05K0A35C", "01H63BQ6A51PQVECDANDCXY8Y8",
		}
		s.Equal(stillDead, ids)

		query := "SELECT id, type, run_at, payload, created_at, " +
			"updated_at, attempts_done, last_attempt_at, last_error " +
			"FROM jobs_queue"
		rows, err := s.db.Query(query)
		s.Require().NoError(err)
		defer rows.Close()

		var alive []worker.Job
		for rows.Next() {
			alive = append(alive, s.scanJob(rows))
		}
		s.Require().NoError(rows.Err())

		ids = ids[:0]
		for _, j := range alive {
			ids = append(ids, j.ID.String())
		}

		s.Equal(resurrectionIDs, ids)
		sample := alive[2]
		s.Equal("index-asset", sample.Type)
		s.JSONEq(
			`{"id":"6d99aab9-dc60-4e92-9c08-8b9e0ad148ea","urn":"urn:firehose:p-godata-id:job:p-godata-id-go-food-pickup-booking-log-bigquery-firehose","type":"job","service":"firehose","name":"p-godata-id-go-food-pickup-booking-log-bigquery-firehose","description":"Migrated from system?","data":{},"url":"","labels":null,"created_at":"2023-07-24T12:39:22.215737Z","updated_at":"2023-07-24T12:39:22.215737Z","version":"0.1","updated_by":{"uuid":"test","email":"test@test.com","provider":""}}`,
			(string)(sample.Payload),
		)
		s.WithinDuration(time.Now(), sample.RunAt, 5*time.Second)
		s.WithinDuration(
			time.Date(2023, time.July, 24, 7, 9, 39, 245731*1000, time.UTC), // 2023-07-24 07:09:22.241788
			sample.CreatedAt,
			time.Millisecond,
		)
		s.WithinDuration(
			time.Date(2023, time.July, 24, 7, 9, 39, 665396*1000, time.UTC), // 2023-07-24 07:09:22.241788
			sample.UpdatedAt,
			time.Millisecond,
		)
		s.Equal(1, sample.AttemptsDone)
		s.WithinDuration(
			time.Date(2023, time.July, 24, 7, 9, 39, 665396*1000, time.UTC), // 2023-07-24 07:09:22.241788
			sample.LastAttemptAt,
			time.Millisecond,
		)
		s.Equal("fail 3", sample.LastError)
	})
}

func (s *ProcessorTestSuite) TestClearDeadJobs() {
	err := testutils.RunMigrations(s.T(), s.db)
	s.Require().NoError(err)

	inserts, err := os.ReadFile("testdata/insert_dead_jobs.sql")
	s.Require().NoError(err)

	_, err = s.db.Exec((string)(inserts))
	s.Require().NoError(err)

	p, err := pgq.NewProcessor(s.ctx, s.testDBConfig())
	s.NoError(err)

	s.Run("InvalidID", func() {
		err := p.ClearDeadJobs(s.ctx, []string{"01H59WDEB4FDS83K8HFNGVJQ7G"})
		s.NoError(err)

		query := "SELECT count(*) FROM dead_jobs"
		var cnt int
		err = s.db.QueryRow(query).Scan(&cnt)
		s.Require().NoError(err)

		s.Equal(12, cnt)
	})

	s.Run("ValidIDs", func() {
		clearIDs := []string{
			"01H63BPCE1JVJHDP1WS9845MN9", "01H63BPVMH6BBQJH776KRJAMH0",
			"01H63BPX1D3S98K5GBADE1Q322", "01H63BQ4GWT8P9BZ8NYZJKMD4W",
			"01H63BPZCB1H6K66M86RY8G85W", "01H63BQ0FR4TJP6431GY0FY3YY",
		}
		err := p.ClearDeadJobs(s.ctx, clearIDs)
		s.NoError(err)

		jobs, err := p.DeadJobs(s.ctx, 20, 0)
		s.NoError(err)

		var ids []string
		for _, j := range jobs {
			ids = append(ids, j.ID.String())
		}
		stillDead := []string{
			"01H63BPY8D4V5SM81KCT5KBW7V", "01H63BQ1HPAT44WS70EG145GKJ",
			"01H63BQ2KVAZEK6DXTQC4D61HP", "01H63BQ3KY6V76FAWAJYVKGDZK",
			"01H63BQ5D0SZEWVX2S05K0A35C", "01H63BQ6A51PQVECDANDCXY8Y8",
		}
		s.Equal(stillDead, ids)

		query := "SELECT count(*) FROM jobs_queue"
		var cnt int
		err = s.db.QueryRow(query).Scan(&cnt)
		s.Require().NoError(err)
		s.Equal(0, cnt)
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
