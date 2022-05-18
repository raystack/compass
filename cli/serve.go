package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/MakeNowJust/heredoc"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/newrelic/go-agent/v3/integrations/nrelasticsearch-v7"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/discussion"
	"github.com/odpf/compass/core/star"
	"github.com/odpf/compass/core/tag"
	"github.com/odpf/compass/core/user"
	compassserver "github.com/odpf/compass/internal/server"
	esStore "github.com/odpf/compass/internal/store/elasticsearch"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/compass/pkg/metrics"
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

	esClient := initElasticsearch(config, logger)
	newRelicMonitor := initNewRelicMonitor(config, logger)
	statsdMonitor := initStatsdMonitor(config, logger)
	pgClient := initPostgres(logger, config)

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

	assetRepository, err := postgres.NewAssetRepository(pgClient, userRepository, 0, config.Service.IdentityProviderDefaultName)
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
		newRelicMonitor.Application(),
		statsdMonitor,
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

func initElasticsearch(config Config, logger log.Logger) *elasticsearch.Client {
	brokers := strings.Split(config.ElasticSearchBrokers, ",")
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: brokers,
		Transport: nrelasticsearch.NewRoundTripper(nil),
		// uncomment below code to debug request and response to elasticsearch
		// Logger: &estransport.ColorLogger{
		//	Output:             os.Stdout,
		//	EnableRequestBody:  true,
		//	EnableResponseBody: true,
		// },
	})
	if err != nil {
		logger.Fatal("error connecting to elasticsearch", "error", err)
	}
	info, err := esInfo(esClient)
	if err != nil {
		logger.Fatal("error obtaining elasticsearch info", "error", err)
	}
	logger.Info("connected to elasticsearch cluster", "config", info)

	return esClient
}

func initPostgres(logger log.Logger, config Config) *postgres.Client {
	pgClient, err := postgres.NewClient(
		postgres.Config{
			Port:     config.DBPort,
			Host:     config.DBHost,
			Name:     config.DBName,
			User:     config.DBUser,
			Password: config.DBPassword,
			SSLMode:  config.DBSSLMode,
		})
	if err != nil {
		logger.Fatal("error creating postgres client", "error", err)
	}
	logger.Info("connected to postgres server", "host", config.DBHost, "port", config.DBPort)

	return pgClient
}

func initNewRelicMonitor(config Config, logger log.Logger) *metrics.NewrelicMonitor {
	if !config.NewRelicEnabled {
		logger.Info("New Relic monitoring is disabled.")
		return nil
	}
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(config.NewRelicAppName),
		newrelic.ConfigLicense(config.NewRelicLicenseKey),
	)
	if err != nil {
		logger.Fatal("unable to create New Relic Application", "error", err)
	}
	logger.Info("New Relic monitoring is enabled for", "config", config.NewRelicAppName)

	monitor := metrics.NewNewrelicMonitor(app)
	return monitor
}

func initStatsdMonitor(config Config, logger log.Logger) *metrics.StatsdMonitor {
	var metricsMonitor *metrics.StatsdMonitor
	if !config.StatsdEnabled {
		logger.Info("statsd metrics monitoring is disabled.")
		return nil
	}
	metricsSeparator := "."
	statsdClient := metrics.NewStatsdClient(config.StatsdAddress)
	metricsMonitor = metrics.NewStatsdMonitor(statsdClient, config.StatsdPrefix, metricsSeparator)
	logger.Info("statsd metrics monitoring is enabled", "statsd address", config.StatsdAddress)

	return metricsMonitor
}

func esInfo(cli *elasticsearch.Client) (string, error) {
	res, err := cli.Info()
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.IsError() {
		return "", errors.New(res.Status())
	}
	var info = struct {
		ClusterName string `json:"cluster_name"`
		Version     struct {
			Number string `json:"number"`
		} `json:"version"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&info)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%q (server version %s)", info.ClusterName, info.Version.Number), nil
}
