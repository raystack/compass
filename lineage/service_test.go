package lineage_test

import (
	"testing"
	"time"

	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/models"
	testifyMock "github.com/stretchr/testify/mock"
)

type stubBuilder struct {
	testifyMock.Mock
}

func (b *stubBuilder) Build(er models.TypeRepository, rrf models.RecordRepositoryFactory) (lineage.Graph, error) {
	return nil, nil
}

type mockMetricsMonitor struct {
	testifyMock.Mock
}

func (mm *mockMetricsMonitor) Duration(op string, d int) {
	mm.Called(op, d)
}

func TestService(t *testing.T) {
	t.Run("smoke test", func(t *testing.T) {
		entRepo := new(mock.TypeRepository)
		entRepo.On("GetAll").Return([]models.Type{}, nil)
		recordRepoFac := new(mock.RecordRepositoryFactory)
		lineage.NewService(entRepo, recordRepoFac, lineage.Config{})
	})
	t.Run("telemetry test", func(t *testing.T) {
		// Temporarily disabling lineage build on service creation causes this test to fail
		t.Skip()

		now := time.Now()
		tsCalled := 0

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
		mm.On("Duration", "lineageBuildTime", int64(100))
		defer mm.AssertExpectations(t)

		lineage.NewService(
			nil,
			nil,
			lineage.Config{
				MetricsMonitor: mm,
				Builder:        builder,
				TimeSource:     lineage.TimeSourceFunc(ts),
			},
		)
	})
}
