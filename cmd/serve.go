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
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/columbus/api"
	"github.com/odpf/columbus/es"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/metrics"
	"github.com/sirupsen/logrus"
)

// Version of the current build. overridden by the build system.
// see "Makefile" for more information
var Version string
var log *logrus.Entry

func Serve() {
	if err := loadConfig(); err != nil {
		panic(err)
	}

	rootLogger := initLogger(config.LogLevel)
	log = rootLogger.WithField("reporter", "main")
	log.Infof("columbus %s starting", Version)

	brokers := strings.Split(config.ElasticSearchBrokers, ",")
	esClient := initElasticsearch(brokers)

	middlewares := []mux.MiddlewareFunc{
		requestLoggerMiddleware(
			rootLogger.WithField("reporter", "http-middleware").Writer(),
		),
	}

	newRelicApp, middlewares := initNewRelic(config, middlewares)
	metricsMonitor, middlewares := initMetricsMonitor(config, middlewares)

	typeRepository := es.NewTypeRepository(esClient)
	recordRepositoryFactory := es.NewRecordRepositoryFactory(esClient)
	recordSearcher, err := es.NewSearcher(esClient, typeWhiteList(config.TypeWhiteListStr))
	if err != nil {
		log.Fatalf("error creating searcher: %v", err)
	}
	lineageService, err := lineage.NewService(typeRepository, recordRepositoryFactory, lineage.Config{
		RefreshInterval: config.LineageRefreshIntervalStr,
		MetricsMonitor:  &metricsMonitor,
	})
	if err != nil {
		log.Fatal(err)
	}

	router := api.NewRouter(api.Config{
		Logger:                  rootLogger,
		RecordSearcher:          recordSearcher,
		RecordRepositoryFactory: recordRepositoryFactory,
		TypeRepository:          typeRepository,
		LineageProvider:         lineageService,
		Middlewares:             middlewares,
	})
	// below handlers still have to be manually wrapped by newrelic core library
	if config.NewRelicEnabled {
		_, router.NotFoundHandler = newrelic.WrapHandle(newRelicApp, "NotFoundHandler", router.NotFoundHandler)
		_, router.MethodNotAllowedHandler = newrelic.WrapHandle(newRelicApp, "MethodNotAllowedHandler", router.MethodNotAllowedHandler)
	}

	serverAddr := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
	log.Printf("starting http server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Errorf("listen and serve: %v", err)
	}
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

func initElasticsearch(brokers []string) *elasticsearch.Client {
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

	return esClient
}

func initNewRelic(config Config, middlewares []mux.MiddlewareFunc) (*newrelic.Application, []mux.MiddlewareFunc) {
	if !config.NewRelicEnabled {
		log.Infof("New Relic monitoring is disabled.")
		return nil, middlewares
	}
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(config.NewRelicAppName),
		newrelic.ConfigLicense(config.NewRelicLicenseKey),
		newrelic.ConfigDebugLogger(os.Stdout),
	)
	if err != nil {
		log.Fatalf("unable to create New Relic Application: %v", err)
	}
	log.Infof("New Relic monitoring is enabled for: %s", config.NewRelicAppName)
	middlewares = append(middlewares, nrgorilla.Middleware(app))
	log.Infof("New relic is setup on the router middleware.")

	return app, middlewares
}

func initMetricsMonitor(config Config, middlewares []mux.MiddlewareFunc) (metrics.Monitor, []mux.MiddlewareFunc) {
	var metricsMonitor metrics.Monitor
	if !config.StatsdEnabled {
		log.Infof("statsd metrics monitoring is disabled.")
		return metricsMonitor, middlewares
	}
	metricsSeparator := "."
	statsdClient := metrics.NewStatsdClient(config.StatsdAddress)
	metricsMonitor = metrics.NewMonitor(statsdClient, config.StatsdPrefix, metricsSeparator)
	middlewares = append(middlewares, telemetryMiddleware(metricsMonitor))
	log.Infof("statsd metrics monitoring is enabled. (%s)", config.StatsdAddress)

	return metricsMonitor, middlewares
}

func requestLoggerMiddleware(dst io.Writer) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return handlers.LoggingHandler(dst, handler)
	}
}

func telemetryMiddleware(mon metrics.Monitor) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return api.MonitoringHandler(handler, mon)
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
