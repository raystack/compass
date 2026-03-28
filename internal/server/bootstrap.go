package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/core/discussion"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/star"
	"github.com/raystack/compass/core/tag"
	"github.com/raystack/compass/core/user"
	"github.com/raystack/compass/internal/config"
	"github.com/raystack/compass/internal/telemetry"
	esStore "github.com/raystack/compass/store/elasticsearch"
	"github.com/raystack/compass/store/postgres"
)

// InitLogger sets up the global slog logger with a JSON handler.
func InitLogger(logLevel string) {
	var level slog.LevelVar
	switch strings.ToLower(logLevel) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn", "warning":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level:     &level,
		AddSource: true,
	}))
	slog.SetDefault(logger)
}

// Start initializes all dependencies and starts the server.
func Start(ctx context.Context, cfg *config.Config, version string) error {
	InitLogger(cfg.LogLevel)
	slog.InfoContext(ctx, "compass starting", "version", version)

	otelCleanup, err := telemetry.Init(ctx, cfg.Telemetry)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	defer otelCleanup()

	esClient, err := initElasticsearch(cfg.Elasticsearch)
	if err != nil {
		return err
	}

	pgClient, err := initPostgres(cfg.DB)
	if err != nil {
		return err
	}
	defer func() {
		slog.Warn("closing db...")
		if err := pgClient.Close(); err != nil {
			slog.Error("error closing db", "error", err)
		}
		slog.Warn("db closed")
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
	userService := user.NewService(userRepository)

	assetRepository, err := postgres.NewAssetRepository(pgClient, userRepository, 0, cfg.Service.Identity.ProviderDefaultName)
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
	namespaceService := namespace.NewService(postgres.NewNamespaceRepository(pgClient), discoveryRepository)

	return Serve(
		ctx,
		cfg.Service,
		namespaceService,
		assetService,
		starService,
		discussionService,
		tagService,
		tagTemplateService,
		userService,
	)
}

// Migrate runs database migrations and creates the default namespace.
func Migrate(ctx context.Context, cfg *config.Config, version string) error {
	InitLogger(cfg.LogLevel)
	slog.InfoContext(ctx, "compass is migrating", "version", version)

	esClient, err := initElasticsearch(cfg.Elasticsearch)
	if err != nil {
		return err
	}

	pgClient, err := initPostgres(cfg.DB)
	if err != nil {
		return err
	}
	defer pgClient.Close()

	ver, err := pgClient.Migrate(cfg.DB)
	if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}

	// create default namespace
	nsService := namespace.NewService(
		postgres.NewNamespaceRepository(pgClient),
		esStore.NewDiscoveryRepository(esClient))
	if _, err = nsService.GetByID(ctx, namespace.DefaultNamespace.ID); errors.Is(err, namespace.ErrNotFound) {
		if _, err := nsService.MigrateDefault(ctx); err != nil {
			return fmt.Errorf("problem with migration %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}
	slog.InfoContext(ctx, "migration finished", "version", ver)
	return nil
}

// MigrateDown rolls back the last database migration.
func MigrateDown(ctx context.Context, cfg *config.Config, version string) error {
	InitLogger(cfg.LogLevel)
	slog.InfoContext(ctx, "compass is rolling back migration", "version", version)

	pgClient, err := initPostgres(cfg.DB)
	if err != nil {
		return err
	}
	defer pgClient.Close()

	ver, err := pgClient.MigrateDown(cfg.DB)
	if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}
	slog.InfoContext(ctx, "migration finished", "version", ver)
	return nil
}

func initElasticsearch(cfg esStore.Config) (*esStore.Client, error) {
	esClient, err := esStore.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new elasticsearch client: %w", err)
	}
	got, err := esClient.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to elasticsearch: %w", err)
	}
	slog.Info("connected to elasticsearch", "info", got)
	return esClient, nil
}

func initPostgres(cfg postgres.Config) (*postgres.Client, error) {
	pgClient, err := postgres.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating postgres client: %w", err)
	}
	slog.Info("connected to postgres server", "host", cfg.Host, "port", cfg.Port)
	return pgClient, nil
}
