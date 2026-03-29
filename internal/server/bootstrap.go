package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/star"
	"github.com/raystack/compass/core/user"
	"github.com/raystack/compass/internal/config"
	compassmcp "github.com/raystack/compass/internal/mcp"
	"github.com/raystack/compass/internal/telemetry"
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

	// init user
	userRepository, err := postgres.NewUserRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create user repository: %w", err)
	}
	userService := user.NewService(userRepository)

	// init star
	starRepository, err := postgres.NewStarRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create star repository: %w", err)
	}
	starService := star.NewService(starRepository)

	// init namespace
	namespaceService := namespace.NewService(postgres.NewNamespaceRepository(pgClient), nil)

	// init entity system (Postgres-native: tsvector + pg_trgm + pgvector)
	entityRepo, err := postgres.NewEntityRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create entity repository: %w", err)
	}
	edgeRepo, err := postgres.NewEdgeRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create edge repository: %w", err)
	}
	entitySearchRepo, err := postgres.NewEntitySearchRepository(pgClient)
	if err != nil {
		return fmt.Errorf("failed to create entity search repository: %w", err)
	}
	entityService := entity.NewService(entityRepo, edgeRepo, entitySearchRepo)

	// init MCP server
	mcpServer := compassmcp.New(entityService, namespace.DefaultNamespace)

	return Serve(
		ctx,
		cfg.Service,
		mcpServer,
		namespaceService,
		starService,
		userService,
		entityService,
		edgeRepo,
	)
}

// Migrate runs database migrations and creates the default namespace.
func Migrate(ctx context.Context, cfg *config.Config, version string) error {
	InitLogger(cfg.LogLevel)
	slog.InfoContext(ctx, "compass is migrating", "version", version)

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
	nsService := namespace.NewService(postgres.NewNamespaceRepository(pgClient), nil)
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

func initPostgres(cfg postgres.Config) (*postgres.Client, error) {
	pgClient, err := postgres.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating postgres client: %w", err)
	}
	slog.Info("connected to postgres server", "host", cfg.Host, "port", cfg.Port)
	return pgClient, nil
}
