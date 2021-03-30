package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/columbus/config"
	"github.com/odpf/columbus/es"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/metrics"
	"github.com/odpf/columbus/web"
	"github.com/sirupsen/logrus"
)

// Version of the current build. overridden by the build system.
// see "Makefile" for more information
var Version string

func Execute() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	rootLogger := initLogger(cfg.LogLevel)

	log := rootLogger.WithField("reporter", "main")
	log.Infof("columbus %s starting", Version)

	brokers := strings.Split(cfg.ElasticSearchBrokers, ",")
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: brokers,
	})
	if err != nil {
		log.Fatalf("error connecting to elasticsearch: %v", err)
	}
	info, err := esInfo(esClient)
	if err != nil {
		log.Fatalf("error obtaining elasticsearch info: %v", err)
	}
	log.Infof("connected to elasticsearch cluster %s", info)

	middlewares := []mux.MiddlewareFunc{
		requestLoggerMiddleware(
			rootLogger.WithField("reporter", "http-middleware").Writer(),
		),
	}

	var newRelicApp *newrelic.Application
	if cfg.NewRelicEnabled {
		newRelicApp = initNewRelic(cfg.NewRelicAppName, cfg.NewRelicLicenseKey)
		log.Infof("New Relic monitoring is enabled for: %s", cfg.NewRelicAppName)

		middlewares = append(middlewares, nrgorilla.Middleware(newRelicApp))
		log.Infof("New relic is setup on the router middleware.")
	} else {
		log.Infof("New Relic monitoring is disabled.")
	}

	var metricsMonitor metrics.Monitor
	if cfg.StatsdEnabled {
		metricsSeparator := "."
		statsdClient := metrics.NewStatsdClient(cfg.StatsdAddress)
		metricsMonitor = metrics.NewMonitor(statsdClient, cfg.StatsdPrefix, metricsSeparator)

		middlewares = append(middlewares, telemetryMiddleware(metricsMonitor))
		log.Infof("statsd metrics monitoring is enabled. (%s)", cfg.StatsdAddress)
	} else {
		log.Infof("statsd metrics monitoring is disabled.")
	}

	typeRepository := es.NewTypeRepository(esClient)
	recordRepositoryFactory := es.NewRecordRepositoryFactory(esClient)
	recordSearcher, err := es.NewSearcher(esClient, typeWhiteList(cfg.TypeWhiteListStr))

	if err != nil {
		log.Fatalf("error creating searcher: %v", err)
	}

	lineageService, err := lineage.NewService(typeRepository, recordRepositoryFactory, lineage.Config{
		RefreshInterval: cfg.LineageRefreshIntervalStr,
		MetricsMonitor:  &metricsMonitor,
	})
	if err != nil {
		log.Fatal(err)
	}

	typeHandler := web.NewTypeHandler(
		rootLogger.WithField("reporter", "type-handler"),
		typeRepository,
		recordRepositoryFactory,
	)
	searchHandler := web.NewSearchHandler(
		rootLogger.WithField("reporter", "search-handler"),
		recordSearcher,
		typeRepository,
	)
	lineageHandler := web.NewLineageHandler(
		rootLogger.WithField("reporter", "lineage-handler"),
		lineageService,
	)

	router := mux.NewRouter()

	// setup routing for different handlers
	router.PathPrefix("/ping").Handler(web.NewHeartbeatHandler())
	router.PathPrefix("/v1/types").Handler(typeHandler)
	router.PathPrefix("/v1/entities").Handler(typeHandler) // For backward compatibility
	router.PathPrefix("/v1/search").Handler(searchHandler)
	router.PathPrefix("/v1/lineage").Handler(lineageHandler)

	// below handlers still have to be manually wrapped by newrelic core library
	if cfg.NewRelicEnabled {
		_, router.NotFoundHandler = newrelic.WrapHandle(newRelicApp, "NotFoundHandler", router.NotFoundHandler)
		_, router.MethodNotAllowedHandler = newrelic.WrapHandle(newRelicApp, "MethodNotAllowedHandler", router.MethodNotAllowedHandler)
	}

	handler := applyMiddlewares(router, middlewares)

	serverAddr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("starting http server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, handler); err != nil {
		log.Errorf("listen and serve: %v", err)
	}
}

func lookupEnvOrString(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func initLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("error parsing log level: %v", err)
	}
	logger.SetOutput(os.Stdout)
	logger.SetLevel(lvl)
	return logger
}

func allLogLevels() []string {
	levels := make([]string, len(logrus.AllLevels))
	for i := 0; i < len(logrus.AllLevels); i++ {
		levels[i] = logrus.AllLevels[i].String()
	}
	return levels
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
	json.NewDecoder(res.Body).Decode(&info)
	return fmt.Sprintf("%q (server version %s)", info.ClusterName, info.Version.Number), nil
}

func typeWhiteList(typeWhiteListStr string) (whiteList []string) {
	indices := strings.Split(typeWhiteListStr, ",")
	for _, index := range indices {
		index = strings.TrimSpace(index)
		if index == "" {
			continue
		}
		whiteList = append(whiteList, index)
	}
	return
}

func initNewRelic(appName, licenseKey string) *newrelic.Application {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(licenseKey),
		newrelic.ConfigDebugLogger(os.Stdout),
	)

	if err != nil {
		log.Fatalf("unable to create New Relic Application: %v", err)
	}

	return app
}

func applyMiddlewares(root *mux.Router, middlewares []mux.MiddlewareFunc) http.Handler {
	for _, middleware := range middlewares {
		root.Use(middleware)
	}
	return root
}

func requestLoggerMiddleware(dst io.Writer) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return handlers.LoggingHandler(dst, handler)
	}
}

func telemetryMiddleware(mon metrics.Monitor) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return web.MonitoringHandler(handler, mon)
	}
}
