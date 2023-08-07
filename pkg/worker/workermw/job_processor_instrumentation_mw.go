package workermw

import (
	"context"
	"sort"
	"time"

	"github.com/goto/compass/pkg/worker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	enqueueDurnHistogram    = "compass.worker.jobs.enqueue.duration"
	dequeueLatencyHistogram = "compass.worker.job.dequeue.latency"
	processDurnHistogram    = "compass.worker.job.process.duration"
)

const (
	attrJobTypes     = attribute.Key("job.types")
	attrJobType      = attribute.Key("job.type")
	attrOpSuccess    = attribute.Key("operation.success")
	attrJobAttemptNo = attribute.Key("job.attempt_number")
	attrJobStatus    = attribute.Key("job.status")
)

type JobProcessorInstrumentation struct {
	next worker.JobProcessor

	enqueueDurn    metric.Float64Histogram
	dequeueLatency metric.Float64Histogram
	processDurn    metric.Float64Histogram
}

func WithJobProcessorInstrumentation() func(worker.JobProcessor) worker.JobProcessor {
	meter := otel.Meter("github.com/goto/compass/pkg/worker/workermw")
	enqueueDurn, err := meter.Float64Histogram(enqueueDurnHistogram)
	handleOtelErr(err)

	dequeueLatency, err := meter.Float64Histogram(dequeueLatencyHistogram)
	handleOtelErr(err)

	processDurn, err := meter.Float64Histogram(processDurnHistogram)
	handleOtelErr(err)

	return func(next worker.JobProcessor) worker.JobProcessor {
		return JobProcessorInstrumentation{
			next:           next,
			enqueueDurn:    enqueueDurn,
			dequeueLatency: dequeueLatency,
			processDurn:    processDurn,
		}
	}
}

func (mw JobProcessorInstrumentation) Enqueue(ctx context.Context, jobs ...worker.Job) (err error) {
	defer func(start time.Time) {
		ms := (float64)(time.Since(start)) / (float64)(time.Millisecond)
		mw.enqueueDurn.Record(ctx, ms, metric.WithAttributes(
			attrJobTypes.StringSlice(jobTypes(jobs)),
			attrOpSuccess.Bool(err == nil),
		))
	}(time.Now())

	return mw.next.Enqueue(ctx, jobs...)
}

func (mw JobProcessorInstrumentation) Process(ctx context.Context, types []string, fn worker.JobExecutorFunc) (err error) {
	start := time.Now()
	wrappedFn := func(ctx context.Context, job worker.Job) (resultJob worker.Job) {
		latency := (float64)(time.Since(job.RunAt)) / (float64)(time.Millisecond)
		mw.dequeueLatency.Record(ctx, latency, metric.WithAttributes(
			attrJobType.String(job.Type),
		))
		defer func() {
			ms := (float64)(time.Since(start)) / (float64)(time.Millisecond)
			mw.processDurn.Record(ctx, ms, metric.WithAttributes(
				attrJobType.String(job.Type),
				attrJobAttemptNo.Int(resultJob.AttemptsDone),
				attrJobStatus.String(jobStatus(resultJob)),
				attrOpSuccess.Bool(resultJob.Status == worker.StatusDone),
			))
		}()
		return fn(ctx, job)
	}
	return mw.next.Process(ctx, types, wrappedFn)
}

func (mw JobProcessorInstrumentation) Stats(ctx context.Context) ([]worker.JobTypeStats, error) {
	return mw.next.Stats(ctx)
}

func jobTypes(jobs []worker.Job) []string {
	var types []string
	for _, j := range jobs {
		types = append(types, j.Type)
	}
	sort.Strings(types)

	return types
}

func jobStatus(j worker.Job) string {
	if j.Status == "" {
		return "retry"
	}

	return (string)(j.Status)
}

func handleOtelErr(err error) {
	if err != nil {
		otel.Handle(err)
	}
}
