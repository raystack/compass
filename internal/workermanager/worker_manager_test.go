package workermanager_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/goto/compass/internal/testutils"
	"github.com/goto/compass/internal/workermanager"
	"github.com/goto/compass/internal/workermanager/mocks"
	"github.com/goto/compass/pkg/worker/pgq"
	"github.com/goto/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func TestNew(t *testing.T) {
	port, err := testutils.RunTestPG(t, log.NewLogrus())
	require.NoError(t, err)

	pgqCfg := pgq.Config{
		Host:     testutils.PGHost,
		Port:     port,
		Name:     testutils.PGName,
		Username: testutils.PGUsername,
		Password: testutils.PGPassword,
	}

	t.Run("InvalidConfig", func(t *testing.T) {
		pgqCfg := pgqCfg
		pgqCfg.Port = 1

		mgr, err := workermanager.New(ctx, workermanager.Deps{
			Config: workermanager.Config{
				WorkerCount:  1,
				PollInterval: time.Second,
				PGQ:          pgqCfg,
			},
		})
		assert.ErrorContains(t, err, "new worker manager: new pgq processor: failed to connect")
		assert.Nil(t, mgr)
	})

	t.Run("Success", func(t *testing.T) {
		mgr, err := workermanager.New(ctx, workermanager.Deps{
			Config: workermanager.Config{
				WorkerCount:  1,
				PollInterval: time.Second,
				PGQ:          pgqCfg,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.NoError(t, mgr.Close())
	})
}

func TestManager_Run(t *testing.T) {
	cases := []struct {
		name        string
		runErr      error
		expectedErr string
	}{
		{name: "Success"},
		{
			name:        "Failure",
			runErr:      errors.New("fail"),
			expectedErr: "fail",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wrkr := mocks.NewWorker(t)
			mgr := workermanager.NewWithWorker(wrkr, workermanager.Deps{})
			wrkr.EXPECT().
				Register("index-asset", mock.AnythingOfType("worker.JobHandler")).
				Return(nil)
			wrkr.EXPECT().
				Register("delete-asset", mock.AnythingOfType("worker.JobHandler")).
				Return(nil)
			wrkr.EXPECT().
				Run(ctx).
				Return(tc.runErr)

			err := mgr.Run(ctx)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
