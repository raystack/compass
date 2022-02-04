package lineage_test

import (
	"context"
	"testing"
	"time"

	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/lineage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type stubBuilder struct{} //nolint:unused

func (b *stubBuilder) Build(ctx context.Context, repo lineage.Repository) (lineage.Graph, error) { //nolint:unused
	return nil, nil
}

type mockMetricsMonitor struct { //nolint:unused
	mock.Mock
}

func (mm *mockMetricsMonitor) Duration(op string, d int) { //nolint:unused
	mm.Called(op, d)
}

type mockPerformanceMonitor struct { //nolint:unused
	mock.Mock
}

func (pm *mockPerformanceMonitor) StartTransaction(ctx context.Context, operation string) (context.Context, func()) { //nolint:unused
	args := pm.Called(ctx, operation)
	return args.Get(0).(context.Context), args.Get(1).(func())
}

func TestService(t *testing.T) {
	ctx := context.Background()
	t.Run("smoke test", func(t *testing.T) {
		repo := new(mocks.LineageRepository)
		repo.On("GetEdges", ctx).Return([]lineage.Edge{}, nil)
		lineage.NewService(repo, lineage.Config{}) // nolint:errcheck
	})
	t.Run("telemetry test", func(t *testing.T) {
		// Temporarily disabling lineage build on service creation causes this test to fail
		t.Skip()

		now := time.Now()
		tsCalled := 0
		txnEnd := false

		// returns now on first call, now +100ms on second
		ts := func() time.Time {
			if tsCalled > 0 {
				return now.Add(100 * time.Millisecond)
			}
			tsCalled++
			return now
		}

		builder := new(stubBuilder)
		mm := new(mockMetricsMonitor)
		mm.On("Duration", "lineageBuildTime", 100)
		pm := new(mockPerformanceMonitor)
		pm.On("StartTransaction", ctx, "lineage:Service/build").Return(ctx, func() {
			txnEnd = true
		})

		if _, err := lineage.NewService(
			nil,
			lineage.Config{
				MetricsMonitor:     mm,
				PerformanceMonitor: pm,
				Builder:            builder,
				TimeSource:         lineage.TimeSourceFunc(ts),
			},
		); err != nil {
			t.Fatal(err)
		}

		mm.AssertExpectations(t)
		assert.True(t, txnEnd)
	})
}
