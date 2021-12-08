package lineage_test

import (
	"context"
	"testing"
	"time"

	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/record"
	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
)

type stubBuilder struct {
	testifyMock.Mock
}

func (b *stubBuilder) Build(ctx context.Context, er record.TypeRepository, rrf discovery.RecordRepositoryFactory) (lineage.Graph, error) {
	return nil, nil
}

type mockMetricsMonitor struct {
	testifyMock.Mock
}

func (mm *mockMetricsMonitor) Duration(op string, d int) {
	mm.Called(op, d)
}

type mockPerformanceMonitor struct {
	testifyMock.Mock
}

func (pm mockPerformanceMonitor) StartTransaction(ctx context.Context, operation string) (context.Context, func()) {
	args := pm.Called(ctx, operation)
	return args.Get(0).(context.Context), args.Get(1).(func())
}

func TestService(t *testing.T) {
	ctx := context.Background()
	t.Run("smoke test", func(t *testing.T) {
		entRepo := new(mock.TypeRepository)
		entRepo.On("GetAll", ctx).Return([]record.Type{}, nil)
		recordRepoFac := new(mock.RecordRepositoryFactory)
		lineage.NewService(entRepo, recordRepoFac, lineage.Config{})
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

		lineage.NewService(
			nil,
			nil,
			lineage.Config{
				MetricsMonitor:     mm,
				PerformanceMonitor: pm,
				Builder:            builder,
				TimeSource:         lineage.TimeSourceFunc(ts),
			},
		)

		mm.AssertExpectations(t)
		assert.True(t, txnEnd)
	})
}
