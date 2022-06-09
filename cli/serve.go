package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/MakeNowJust/heredoc"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/discussion"
	"github.com/odpf/compass/core/star"
	"github.com/odpf/compass/core/tag"
	"github.com/odpf/compass/core/user"
	compassserver "github.com/odpf/compass/internal/server"
	esStore "github.com/odpf/compass/internal/store/elasticsearch"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/compass/pkg/statsd"
	"github.com/odpf/salt/log"
	"github.com/spf13/cobra"
)

// Version of the current build. overridden by the build system.
// see "Makefile" for more information
var (
	Version string
)

func cmdServe() *cobra.Command {
	return &cobra.Command{
		Use:     "serve",
		Short:   "Serve gRPC & HTTP service",
		Long:    heredoc.Doc(`Serve gRPC & HTTP on a port defined in PORT env var.`),
		Aliases: []string{"server", "start"},
		Example: heredoc.Doc(`
			$ compass serve
		`),
		Args: cobra.NoArgs,
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}
			return runServer(cfg)
		},
	}
}

func runServer(config Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := initLogger(config.LogLevel)
	logger.Info("compass starting", "version", Version)

	nrApp, err := initNewRelicMonitor(config, logger)
	if err != nil {
		return err
	}
	statsdReporter, err := statsd.Init(logger, config.StatsD)
	if err != nil {
		return err
	}

	esClient, err := initElasticsearch(logger, config.Elasticsearch)
	if err != nil {
		return err
	}

	pgClient, err := initPostgres(logger, config)
	if err != nil {
		return err
	}

	// init tag
	tagRepository, err := postgres.NewTagRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create new tag repository: %w", err)
	}
	tagTemplateRepository, err := postgres.NewTagTemplateRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create new tag template repository: %w", err)
	}
	tagTemplateService := tag.NewTemplateService(tagTemplateRepository)
	tagService := tag.NewService(tagRepository, tagTemplateService)

	// init user
	userRepository, err := postgres.NewUserRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create new user repository: %w", err)
	}
	userService := user.NewService(logger, userRepository, user.ServiceWithStatsDReporter(statsdReporter))

	assetRepository, err := postgres.NewAssetRepository(pgClient, userRepository, 0, config.Service.Identity.ProviderDefaultName)
	if err != nil {
		return fmt.Errorf("failed to create new asset repository: %w", err)
	}
	discoveryRepository := esStore.NewDiscoveryRepository(esClient)
	lineageRepository, err := postgres.NewLineageRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create new lineage repository: %w", err)
	}
	assetService := asset.NewService(assetRepository, discoveryRepository, lineageRepository)

	// init discussion
	discussionRepository, err := postgres.NewDiscussionRepository(pgClient, 0)
	if err != nil {
		return fmt.Errorf("failed to create new discussion repository: %w", err)
	}
	discussionService := discussion.NewService(discussionRepository)

	// init star
	starRepository, err := postgres.NewStarRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create new star repository: %w", err)
	}
	starService := star.NewService(starRepository)

	return compassserver.Serve(
		ctx,
		config.Service,
		logger,
		pgClient,
		nrApp,
		statsdReporter,
		assetService,
		starService,
		discussionService,
		tagService,
		tagTemplateService,
		userService,
	)
}

func initLogger(logLevel string) *log.Logrus {
	logger := log.NewLogrus(
		log.LogrusWithLevel(logLevel),
		log.LogrusWithWriter(os.Stdout),
	)
	return logger
}

func initElasticsearch(logger log.Logger, config esStore.Config) (*esStore.Client, error) {
	esClient, err := esStore.NewClient(logger, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create new elasticsearch client: %w", err)
	}
	got, err := esClient.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to elasticsearch: %w", err)
	}
	logger.Info("connected to elasticsearch", "info", got)
	return esClient, nil
}

func initPostgres(logger log.Logger, config Config) (*postgres.Client, error) {
	pgClient, err := postgres.NewClient(config.DB)
	if err != nil {
		return nil, fmt.Errorf("error creating postgres client: %w", err)
	}
	logger.Info("connected to postgres server", "host", config.DB.Host, "port", config.DB.Port)

	return pgClient, nil
}

func initNewRelicMonitor(config Config, logger log.Logger) (*newrelic.Application, error) {
	if !config.NewRelic.Enabled {
		logger.Info("New Relic monitoring is disabled.")
		return nil, nil
	}
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(config.NewRelic.AppName),
		newrelic.ConfigLicense(config.NewRelic.LicenseKey),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create New Relic Application: %w", err)
	}
	logger.Info("New Relic monitoring is enabled for", "config", config.NewRelic.AppName)

	return app, nil
}
