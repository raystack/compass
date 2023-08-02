package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/goto/compass/internal/store/elasticsearch"
	"github.com/goto/compass/internal/workermanager"
	"github.com/goto/compass/pkg/telemetry"
	"github.com/spf13/cobra"
)

func workerCmd(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worker <command>",
		Aliases: []string{"s"},
		Short:   "Run compass worker",
		Long:    "Worker management commands.",
		Example: heredoc.Doc(`
			$ compass worker start
			$ compass worker start -c ./config.yaml
		`),
	}

	cmd.AddCommand(workerStartCommand(cfg))

	return cmd
}

func workerStartCommand(cfg *Config) *cobra.Command {
	c := &cobra.Command{
		Use:     "start",
		Short:   "Start worker to start processing jobs and a web server for dead job management on default port 8085",
		Example: "compass worker start",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runWorker(cmd.Context(), cfg); err != nil {
				return fmt.Errorf("run worker: %w", err)
			}
			return nil
		},
	}

	return c
}

func runWorker(ctx context.Context, cfg *Config) error {
	if !cfg.Worker.Enabled {
		return errors.New("worker is disabled")
	}

	logger := initLogger(cfg.LogLevel)
	logger.Info("Compass worker starting", "version", Version)

	_, cleanUp, err := telemetry.Init(ctx, cfg.Telemetry, logger)
	if err != nil {
		return err
	}

	defer cleanUp()

	esClient, err := initElasticsearch(logger, cfg.Elasticsearch)
	if err != nil {
		return err
	}

	mgr, err := workermanager.New(ctx, workermanager.Deps{
		Config:        cfg.Worker,
		DiscoveryRepo: elasticsearch.NewDiscoveryRepository(esClient, logger),
		Logger:        logger,
	})
	if err != nil {
		return err
	}

	defer func() {
		if err := mgr.Close(); err != nil {
			logger.Error("Close worker manager", "err", err)
		}
	}()

	return mgr.Run(ctx)
}
