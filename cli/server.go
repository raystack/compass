package cli

import (
	"context"
	"fmt"
	"github.com/odpf/compass/core/namespace"
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

func serverCmd(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server <command>",
		Aliases: []string{"s"},
		Short:   "Run compass server",
		Long:    "Server management commands.",
		Example: heredoc.Doc(`
			$ compass server start
			$ compass server start -c ./config.yaml
			$ compass server migrate
			$ compass server migrate -c ./config.yaml
		`),
	}

	cmd.AddCommand(
		serverStartCommand(cfg),
		serverMigrateCommand(cfg),
	)

	return cmd
}

func serverStartCommand(cfg *Config) *cobra.Command {

	c := &cobra.Command{
		Use:     "start",
		Short:   "Start server on default port 8080",
		Example: "compass server start",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cfg)
		},
	}

	return c
}

func serverMigrateCommand(cfg *Config) *cobra.Command {
	var down bool
	c := &cobra.Command{
		Use:   "migrate",
		Short: "Run storage migration",
		Example: heredoc.Doc(`
			$ compass server migrate
			$ compass server migrate --down
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			if down {
				return migrateDown(cfg)
			}
			return runMigrations(ctx, cfg)
		},
	}
	c.Flags().BoolVar(&down, "down", false, "rollback migration one step")
	return c
}

func runServer(config *Config) error {
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

	// init namespace
	namespaceService := namespace.NewService(logger, postgres.NewNamespaceRepository(pgClient), discoveryRepository)

	return compassserver.Serve(
		ctx,
		config.Service,
		logger,
		pgClient,
		nrApp,
		statsdReporter,
		namespaceService,
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

func initPostgres(logger log.Logger, config *Config) (*postgres.Client, error) {
	pgClient, err := postgres.NewClient(config.DB)
	if err != nil {
		return nil, fmt.Errorf("error creating postgres client: %w", err)
	}
	logger.Info("connected to postgres server", "host", config.DB.Host, "port", config.DB.Port)

	return pgClient, nil
}

func initNewRelicMonitor(config *Config, logger log.Logger) (*newrelic.Application, error) {
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

func runMigrations(ctx context.Context, config *Config) error {
	fmt.Println("Preparing migration...")

	logger := initLogger(config.LogLevel)
	logger.Info("compass is migrating", "version", Version)

	logger.Info("Migrating Postgres & ElasticSearch...")
	esClient, err := initElasticsearch(logger, config.Elasticsearch)
	if err != nil {
		return err
	}

	logger.Info("Initiating Postgres client...")
	pgClient, err := postgres.NewClient(config.DB)
	if err != nil {
		logger.Error("failed to prepare migration", "error", err)
		return err
	}

	ver, err := pgClient.Migrate(config.DB)
	if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}

	// create default namespace
	nsService := namespace.NewService(logger,
		postgres.NewNamespaceRepository(pgClient),
		esStore.NewDiscoveryRepository(esClient))
	if _, err = nsService.GetByID(ctx, namespace.DefaultNamespace.ID); err == postgres.ErrNamespaceNotFound {
		// create default
		if _, err := nsService.MigrateDefault(ctx); err != nil {
			return fmt.Errorf("problem with migration %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}
	logger.Info(fmt.Sprintf("Migration finished. Version: %d", ver))
	return nil
}

func migrateDown(config *Config) error {
	fmt.Println("Preparing rolling back one step of migration...")

	logger := initLogger(config.LogLevel)
	logger.Info("compass is migrating", "version", Version)

	logger.Info("Initiating Postgres client...")
	pgClient, err := postgres.NewClient(config.DB)
	if err != nil {
		logger.Error("failed to prepare migration", "error", err)
		return err
	}

	ver, err := pgClient.MigrateDown(config.DB)
	if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}
	logger.Info(fmt.Sprintf("Migration finished. Version: %d", ver))
	return nil
}
