package worker

import (
	"context"
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//go:embed dead_jobs_page.html
var deadJobsPageTplSrc string

//go:generate mockery --name=DeadJobManager -r --case underscore --with-expecter --structname DeadJobManager --filename dead_job_manager_mock.go --output=./mocks

type DeadJobManager interface {
	DeadJobs(ctx context.Context, size, offset int) ([]Job, error)
	Resurrect(ctx context.Context, jobIDs []string) error
	ClearDeadJobs(ctx context.Context, jobIDs []string) error
}

const (
	listDeadJobsPath  = "/dead-jobs"
	resurrectJobsPath = "/resurrect-jobs"
	clearJobsPath     = "/clear-jobs"
)

// DeadJobManagementHandler returns a http handler with endpoints for dead job
// management:
//   - /dead-jobs: JSON/HTML response with content negotiation. Response is
//     paginated.
//   - /resurrect-jobs: Move specified dead jobs to jobs_queue table.
//   - /clear-jobs: Remove specified dead jobs from dead_jobs table.
func DeadJobManagementHandler(mgr DeadJobManager) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(
		listDeadJobsPath,
		otelhttp.NewMiddleware("list_dead_jobs")(
			otelhttp.WithRouteTag(listDeadJobsPath, deadJobsHandler(mgr)),
		),
	)
	mux.Handle(
		resurrectJobsPath,
		otelhttp.NewMiddleware("resurrect_jobs")(
			otelhttp.WithRouteTag(resurrectJobsPath, jobsTransformerHandler(func(ctx context.Context, jobIDs []string) error {
				return mgr.Resurrect(ctx, jobIDs)
			})),
		),
	)
	mux.Handle(
		clearJobsPath,
		otelhttp.NewMiddleware("clear_jobs")(
			otelhttp.WithRouteTag(clearJobsPath, jobsTransformerHandler(func(ctx context.Context, jobIDs []string) error {
				return mgr.ClearDeadJobs(ctx, jobIDs)
			})),
		),
	)
	return mux
}

func deadJobsHandler(mgr DeadJobManager) http.HandlerFunc {
	deadJobsPageTpl := template.Must(template.New("").Parse(deadJobsPageTplSrc))

	return func(w http.ResponseWriter, r *http.Request) {
		qry := r.URL.Query()
		size, err := strconv.Atoi(qry.Get("size"))
		if err != nil || size <= 0 {
			size = 20
		}

		offset, _ := strconv.Atoi(qry.Get("offset"))
		if offset <= 0 {
			offset = 0
		}

		jobs, err := mgr.DeadJobs(r.Context(), size, offset)
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, err)
			return
		}

		if strings.Contains(r.Header.Get("Accept"), "application/json") {
			writeJSONResponse(w, http.StatusOK, jobs)
			return
		}

		tplData := map[string]any{
			"jobs":      jobs,
			"page_size": size,
			"next_page": offset + size,
			"prev_page": offset - size,
			"err_msg":   strings.TrimSpace(qry.Get("error")),
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		if err := deadJobsPageTpl.Execute(w, tplData); err != nil {
			sendTo := listDeadJobsPath + "?error=" + url.QueryEscape(err.Error())
			http.Redirect(w, r, sendTo, http.StatusSeeOther)
			return
		}
	}
}

func jobsTransformerHandler(fn func(context.Context, []string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			sendTo := listDeadJobsPath + "?error=" + url.QueryEscape(err.Error())
			http.Redirect(w, r, sendTo, http.StatusSeeOther)
			return
		}

		jobIDs := r.Form["job_ids"]
		if len(jobIDs) == 0 {
			sendTo := listDeadJobsPath + "?error=" + url.QueryEscape("no job IDs specified")
			http.Redirect(w, r, sendTo, http.StatusSeeOther)
			return
		}

		if err := fn(r.Context(), jobIDs); err != nil {
			sendTo := listDeadJobsPath + "?error=" + url.QueryEscape(err.Error())
			http.Redirect(w, r, sendTo, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, listDeadJobsPath, http.StatusSeeOther)
	}
}

func writeJSONResponse(w http.ResponseWriter, status int, v any) {
	if err, ok := v.(error); ok {
		v = map[string]interface{}{"error": err.Error()}
	}

	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "encode response failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
