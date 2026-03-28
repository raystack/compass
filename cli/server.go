package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/raystack/compass/core/namespace"
	"os"
	"os/signal"
	"syscall"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/core/discussion"
	"github.com/raystack/compass/core/star"
	"github.com/raystack/compass/core/tag"
	"github.com/raystack/compass/core/user"
	compassserver "github.com/raystack/compass/internal/server"
	esStore "github.com/raystack/compass/store/elasticsearch"
	"github.com/raystack/compass/store/postgres"
	"github.com/raystack/compass/internal/telemetry"
	log "github.com/raystack/salt/observability/logger"
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
				return migrateDown(ctx, cfg)
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

	otelCleanup, err := telemetry.Init(ctx, config.Telemetry, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	defer otelCleanup()

	esClient, err := initElasticsearch(logger, config.Elasticsearch)
	if err != nil {
		return err
	}

	pgClient, err := initPostgres(logger, config)
	if err != nil {
		return err
	}
	defer func() {
		logger.Warn("closing db...")
		if err := pgClient.Close(); err != nil {
			logger.Error("error when closing db", "err", err)
		}
		logger.Warn("db closed...")
	}()

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
	userService := user.NewService(logger, userRepository)

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

func runMigrations(ctx context.Context, config *Config) error {
	logger := initLogger(config.LogLevel)
	logger.Info("compass is migrating", "version", Version)

	esClient, err := initElasticsearch(logger, config.Elasticsearch)
	if err != nil {
		return err
	}

	pgClient, err := initPostgres(logger, config)
	if err != nil {
		return err
	}
	defer pgClient.Close()

	ver, err := pgClient.Migrate(config.DB)
	if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}

	// create default namespace
	nsService := namespace.NewService(logger,
		postgres.NewNamespaceRepository(pgClient),
		esStore.NewDiscoveryRepository(esClient))
	if _, err = nsService.GetByID(ctx, namespace.DefaultNamespace.ID); errors.Is(err, namespace.ErrNotFound) {
		// create default
		if _, err := nsService.MigrateDefault(ctx); err != nil {
			return fmt.Errorf("problem with migration %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}
	logger.Info("migration finished", "version", ver)
	return nil
}

func migrateDown(ctx context.Context, config *Config) error {
	logger := initLogger(config.LogLevel)
	logger.Info("compass is rolling back migration", "version", Version)

	pgClient, err := initPostgres(logger, config)
	if err != nil {
		return err
	}
	defer pgClient.Close()

	ver, err := pgClient.MigrateDown(config.DB)
	if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}
	logger.Info("migration finished", "version", ver)
	return nil
}
