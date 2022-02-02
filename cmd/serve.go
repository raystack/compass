package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	nrelasticsearch "github.com/newrelic/go-agent/v3/integrations/nrelasticsearch-v7"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/columbus/api"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/metrics"
	esStore "github.com/odpf/columbus/store/elasticsearch"
	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/tag"
	"github.com/odpf/salt/log"
)

// Version of the current build. overridden by the build system.
// see "Makefile" for more information
var Version string

func Serve() {
	if err := loadConfig(); err != nil {
		panic(err)
	}

	logger := initLogger(config.LogLevel)
	logger.Info("columbus starting", "version", Version)

	esClient := initElasticsearch(config, logger)
	newRelicMonitor := initNewRelicMonitor(config, logger)
	statsdMonitor := initStatsdMonitor(config, logger)
	router := initRouter(esClient, newRelicMonitor, statsdMonitor, logger)

	serverAddr := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
	logger.Info(fmt.Sprintf("starting http server on %s", serverAddr))
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		logger.Error("listen and serve", "error", err)
	}
}

func initRouter(
	esClient *elasticsearch.Client,
	nrMonitor *metrics.NewrelicMonitor,
	statsdMonitor *metrics.StatsdMonitor,
	logger log.Logger,
) *mux.Router {
	typeRepository := esStore.NewTypeRepository(esClient)
	recordRepositoryFactory := esStore.NewRecordRepositoryFactory(esClient)
	recordSearcher, err := esStore.NewSearcher(esStore.SearcherConfig{
		Client: esClient,
	})
	if err != nil {
		logger.Fatal("error creating searcher", "error", err)
	}

	pgClient := initPostgres(logger, config)
	tagRepository, err := postgres.NewTagRepository(pgClient)
	if err != nil {
		logger.Fatal("failed to create new tag repository", "error", err)
	}
	tagTemplateRepository, err := postgres.NewTagTemplateRepository(pgClient)
	if err != nil {
		logger.Fatal("failed to create new tag template repository", "error", err)
	}
	tagTemplateService := tag.NewTemplateService(tagTemplateRepository)
	tagService := tag.NewService(
		tagRepository,
		tagTemplateService,
	)

	userRepository, err := postgres.NewUserRepository(pgClient)
	if err != nil {
		logger.Fatal("failed to create new user repository", "error", err)
	}
	assetRepository, err := postgres.NewAssetRepository(pgClient, userRepository, 0)
	if err != nil {
		logger.Fatal("failed to create new asset repository", "error", err)
	}

	discoveryRepo := esStore.NewDiscoveryRepository(esClient)
	lineageRepo, err := postgres.NewLineageRepository(pgClient)
	if err != nil {
		logger.Fatal("failed to create new lineage repository", "error", err)
	}
	lineageService, err := lineage.NewService(lineageRepo, lineage.Config{
		RefreshInterval:    config.LineageRefreshIntervalStr,
		MetricsMonitor:     statsdMonitor,
		PerformanceMonitor: nrMonitor,
	})
	if err != nil {
		logger.Fatal("failed to create service", "error", err)
	}
	// build lineage asynchronously
	go func() {
		lineageService.ForceBuild()
		logger.Info("lineage build complete")
	}()

	router := mux.NewRouter()
	if nrMonitor != nil {
		nrMonitor.MonitorRouter(router)
	}
	if statsdMonitor != nil {
		statsdMonitor.MonitorRouter(router)
	}
	router.Use(requestLoggerMiddleware(
		logger.Writer(),
	))

	api.RegisterRoutes(router, api.Config{
		Logger:                  logger,
		AssetRepository:         assetRepository,
		DiscoveryRepository:     discoveryRepo,
		TypeRepository:          typeRepository,
		DiscoveryService:        discovery.NewService(recordRepositoryFactory, recordSearcher),
		RecordRepositoryFactory: recordRepositoryFactory,
		LineageProvider:         lineageService,
		TagService:              tagService,
		TagTemplateService:      tagTemplateService,
	})

	return router
}

func initLogger(logLevel string) log.Logger {
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
		// 	Output:             os.Stdout,
		// 	EnableRequestBody:  true,
		// 	EnableResponseBody: true,
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
	logger.Info("connected to postgres server %s:%d", "host", config.DBHost, "port", config.DBPort)

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

func requestLoggerMiddleware(dst io.Writer) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return handlers.LoggingHandler(dst, handler)
	}
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
