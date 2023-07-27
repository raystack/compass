package worker_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/goto/compass/internal/testutils"
	"github.com/goto/compass/pkg/worker"
	"github.com/goto/compass/pkg/worker/mocks"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func TestDeadJobManagementHandler(t *testing.T) {
	frozenTime := time.Unix(1654082526, 0).UTC()
	job := worker.Job{
		ID: ulid.MustParse("01H65SPDDWGN753S5W8S0YHYXD"),
		JobSpec: worker.JobSpec{
			Type:    "test",
			Payload: []byte("payload 1, 2, 3"),
			RunAt:   frozenTime,
		},
		CreatedAt:     frozenTime,
		UpdatedAt:     frozenTime,
		AttemptsDone:  1,
		Status:        "",
		LastAttemptAt: frozenTime,
		LastError:     "fail",
	}

	t.Run("DeadJobsPage", func(t *testing.T) {
		t.Run("AcceptJSON", func(t *testing.T) {
			cases := []struct {
				name     string
				query    string
				setup    func(mgr *mocks.DeadJobManager)
				expected response
			}{
				{
					name:  "ValidSizeAndOffset",
					query: "size=10&offset=1",
					setup: func(mgr *mocks.DeadJobManager) {
						mgr.EXPECT().DeadJobs(testutils.OfTypeContext(), 10, 1).
							Return([]worker.Job{job}, nil)
					},
					expected: response{
						Status:  http.StatusOK,
						Headers: map[string]string{"Content-Type": "application/json"},
						Body:    `[{"id":"01H65SPDDWGN753S5W8S0YHYXD","type":"test","args":"cGF5bG9hZCAxLCAyLCAz","run_at":"2022-06-01T11:22:06Z","created_at":"2022-06-01T11:22:06Z","updated_at":"2022-06-01T11:22:06Z","attempts_done":1,"last_error":"fail","last_attempt_at":"2022-06-01T11:22:06Z"}]`,
					},
				},
				{
					name: "WithoutSizeAndOffset",
					setup: func(mgr *mocks.DeadJobManager) {
						mgr.EXPECT().DeadJobs(testutils.OfTypeContext(), 20, 0).
							Return([]worker.Job{}, nil)
					},
					expected: response{
						Status:  http.StatusOK,
						Headers: map[string]string{"Content-Type": "application/json"},
						Body:    `[]`,
					},
				},
				{
					name:  "WithInvalidSizeAndOffset",
					query: "size=-1&offset=-1",
					setup: func(mgr *mocks.DeadJobManager) {
						mgr.EXPECT().DeadJobs(testutils.OfTypeContext(), 20, 0).
							Return([]worker.Job{}, nil)
					},
					expected: response{
						Status:  http.StatusOK,
						Headers: map[string]string{"Content-Type": "application/json"},
						Body:    `[]`,
					},
				},
				{
					name: "DeadJobManagerError",
					setup: func(mgr *mocks.DeadJobManager) {
						mgr.EXPECT().DeadJobs(testutils.OfTypeContext(), 20, 0).
							Return(nil, errors.New("boiled a broken egg"))
					},
					expected: response{
						Status:  http.StatusInternalServerError,
						Headers: map[string]string{"Content-Type": "application/json"},
						Body:    `{"error":"boiled a broken egg"}`,
					},
				},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					mgr := mocks.NewDeadJobManager(t)
					tc.setup(mgr)

					req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/dead-jobs?"+tc.query, nil)
					req.Header.Set("Accept", "application/json")

					res := recordResponse(worker.DeadJobManagementHandler(mgr), req)
					defer res.Body.Close()

					matchResponse(t, tc.expected, res)
				})
			}
		})

		t.Run("AcceptTextHTML", func(t *testing.T) {
			mgr := mocks.NewDeadJobManager(t)
			mgr.EXPECT().DeadJobs(testutils.OfTypeContext(), 20, 0).
				Return([]worker.Job{job}, nil)

			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/dead-jobs", nil)
			req.Header.Set("Accept", "text/html")

			res := recordResponse(worker.DeadJobManagementHandler(mgr), req)
			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)
			data, _ := io.ReadAll(res.Body)
			html := (string)(data)
			assert.Contains(t, html, "test / 01H65SPDDWGN753S5W8S0YHYXD")
			assert.Contains(t, html, "payload 1, 2, 3")
			assert.Contains(t, html, "Last Error: fail")
		})
	})

	t.Run("ResurrectJobs", func(t *testing.T) {
		cases := []struct {
			name     string
			setup    func(mgr *mocks.DeadJobManager)
			body     io.Reader
			expected response
		}{
			{
				name:  "MalformedFormPost",
				body:  nil,
				setup: func(mgr *mocks.DeadJobManager) {},
				expected: response{
					Status: http.StatusSeeOther,
					Headers: map[string]string{
						"Location": "/dead-jobs?error=" + url.QueryEscape("missing form body"),
					},
				},
			},
			{
				name:  "MissingJobIDs",
				body:  strings.NewReader(""),
				setup: func(mgr *mocks.DeadJobManager) {},
				expected: response{
					Status: http.StatusSeeOther,
					Headers: map[string]string{
						"Location": "/dead-jobs?error=" + url.QueryEscape("no job IDs specified"),
					},
				},
			},
			{
				name: "DeadJobManagerError",
				body: strings.NewReader(url.Values{
					"job_ids": []string{"01H65SPDDWGN753S5W8S0YHYXD", "01H63BQ5D0SZEWVX2S05K0A35C"},
				}.Encode()),
				setup: func(mgr *mocks.DeadJobManager) {
					mgr.EXPECT().Resurrect(testutils.OfTypeContext(), []string{"01H65SPDDWGN753S5W8S0YHYXD", "01H63BQ5D0SZEWVX2S05K0A35C"}).
						Return(errors.New("technical debt"))
				},
				expected: response{
					Status: http.StatusSeeOther,
					Headers: map[string]string{
						"Location": "/dead-jobs?error=" + url.QueryEscape("technical debt"),
					},
				},
			},
			{
				name: "Success",
				body: strings.NewReader(url.Values{
					"job_ids": []string{"01H65SPDDWGN753S5W8S0YHYXD", "01H63BQ5D0SZEWVX2S05K0A35C"},
				}.Encode()),
				setup: func(mgr *mocks.DeadJobManager) {
					mgr.EXPECT().Resurrect(testutils.OfTypeContext(), []string{"01H65SPDDWGN753S5W8S0YHYXD", "01H63BQ5D0SZEWVX2S05K0A35C"}).
						Return(nil)
				},
				expected: response{
					Status: http.StatusSeeOther,
					Headers: map[string]string{
						"Location": "/dead-jobs",
					},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				mgr := mocks.NewDeadJobManager(t)
				tc.setup(mgr)

				req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/resurrect-jobs", tc.body)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				res := recordResponse(worker.DeadJobManagementHandler(mgr), req)
				defer res.Body.Close()

				matchResponse(t, tc.expected, res)
			})
		}
	})

	t.Run("ClearDeadJobs", func(t *testing.T) {
		cases := []struct {
			name     string
			setup    func(mgr *mocks.DeadJobManager)
			body     io.Reader
			expected response
		}{
			{
				name:  "MalformedFormPost",
				body:  nil,
				setup: func(mgr *mocks.DeadJobManager) {},
				expected: response{
					Status: http.StatusSeeOther,
					Headers: map[string]string{
						"Location": "/dead-jobs?error=" + url.QueryEscape("missing form body"),
					},
				},
			},
			{
				name:  "MissingJobIDs",
				body:  strings.NewReader(""),
				setup: func(mgr *mocks.DeadJobManager) {},
				expected: response{
					Status: http.StatusSeeOther,
					Headers: map[string]string{
						"Location": "/dead-jobs?error=" + url.QueryEscape("no job IDs specified"),
					},
				},
			},
			{
				name: "DeadJobManagerError",
				body: strings.NewReader(url.Values{
					"job_ids": []string{"01H65SPDDWGN753S5W8S0YHYXD", "01H63BQ5D0SZEWVX2S05K0A35C"},
				}.Encode()),
				setup: func(mgr *mocks.DeadJobManager) {
					mgr.EXPECT().ClearDeadJobs(testutils.OfTypeContext(), []string{"01H65SPDDWGN753S5W8S0YHYXD", "01H63BQ5D0SZEWVX2S05K0A35C"}).
						Return(errors.New("technical debt"))
				},
				expected: response{
					Status: http.StatusSeeOther,
					Headers: map[string]string{
						"Location": "/dead-jobs?error=" + url.QueryEscape("technical debt"),
					},
				},
			},
			{
				name: "Success",
				body: strings.NewReader(url.Values{
					"job_ids": []string{"01H65SPDDWGN753S5W8S0YHYXD", "01H63BQ5D0SZEWVX2S05K0A35C"},
				}.Encode()),
				setup: func(mgr *mocks.DeadJobManager) {
					mgr.EXPECT().ClearDeadJobs(testutils.OfTypeContext(), []string{"01H65SPDDWGN753S5W8S0YHYXD", "01H63BQ5D0SZEWVX2S05K0A35C"}).
						Return(nil)
				},
				expected: response{
					Status: http.StatusSeeOther,
					Headers: map[string]string{
						"Location": "/dead-jobs",
					},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				mgr := mocks.NewDeadJobManager(t)
				tc.setup(mgr)

				req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/clear-jobs", tc.body)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				res := recordResponse(worker.DeadJobManagementHandler(mgr), req)
				defer res.Body.Close()

				matchResponse(t, tc.expected, res)
			})
		}
	})
}

func recordResponse(h http.Handler, req *http.Request) *http.Response {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Result()
}

type response struct {
	Status  int
	Headers map[string]string
	Body    string
}

func matchResponse(t *testing.T, expected response, actual *http.Response) {
	t.Helper()

	assert.Equal(t, expected.Status, actual.StatusCode)

	for k, v := range expected.Headers {
		assert.Equal(t, v, actual.Header.Get(k))
	}

	body, err := io.ReadAll(actual.Body)
	if !assert.NoError(t, err) {
		return
	}

	if expected.Body == "" {
		assert.Empty(t, body)
		return
	}

	assert.JSONEq(t, expected.Body, (string)(body))
}
