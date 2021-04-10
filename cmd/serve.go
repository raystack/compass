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
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/metrics"
	"github.com/odpf/columbus/store"
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

	esClient := initElasticsearch(config)
	newRelicMonitor := initNewRelicMonitor(config)
	statsdMonitor := initStatsdMonitor(config)
	router := initRouter(esClient, newRelicMonitor, statsdMonitor, rootLogger)

	serverAddr := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
	log.Printf("starting http server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Errorf("listen and serve: %v", err)
	}
}

func initRouter(
	esClient *elasticsearch.Client,
	nrMonitor *metrics.NewrelicMonitor,
	statsdMonitor *metrics.StatsdMonitor,
	rootLogger logrus.FieldLogger,
) *mux.Router {
	typeRepository := store.NewTypeRepository(esClient)
	recordRepositoryFactory := store.NewRecordRepositoryFactory(esClient)
	recordSearcher, err := store.NewSearcher(store.SearcherConfig{
		Client:                 esClient,
		TypeRepo:               typeRepository,
		TypeWhiteList:          typeWhiteList(config.TypeWhiteListStr),
		CachedTypesMapDuration: config.SearchTypesCacheDuration,
	})
	if err != nil {
		log.Fatalf("error creating searcher: %v", err)
	}
	lineageService, err := lineage.NewService(typeRepository, recordRepositoryFactory, lineage.Config{
		RefreshInterval: config.LineageRefreshIntervalStr,
		MetricsMonitor:  statsdMonitor,
	})
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	if nrMonitor != nil {
		nrMonitor.MonitorRouter(router)
	}
	if statsdMonitor != nil {
		statsdMonitor.MonitorRouter(router)
	}
	router.Use(requestLoggerMiddleware(
		rootLogger.WithField("reporter", "http-middleware").Writer(),
	))
	api.RegisterRoutes(router, api.Config{
		Logger:                  rootLogger,
		RecordSearcher:          recordSearcher,
		RecordRepositoryFactory: recordRepositoryFactory,
		TypeRepository:          typeRepository,
		LineageProvider:         lineageService,
	})

	return router
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

func initElasticsearch(config Config) *elasticsearch.Client {
	brokers := strings.Split(config.ElasticSearchBrokers, ",")
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: brokers,
		Transport: nrelasticsearch.NewRoundTripper(nil),
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

func initNewRelicMonitor(config Config) *metrics.NewrelicMonitor {
	if !config.NewRelicEnabled {
		log.Infof("New Relic monitoring is disabled.")
		return nil
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

	monitor := metrics.NewNewrelicMonitor(app)
	return monitor
}

func initStatsdMonitor(config Config) *metrics.StatsdMonitor {
	var metricsMonitor *metrics.StatsdMonitor
	if !config.StatsdEnabled {
		log.Infof("statsd metrics monitoring is disabled.")
		return nil
	}
	metricsSeparator := "."
	statsdClient := metrics.NewStatsdClient(config.StatsdAddress)
	metricsMonitor = metrics.NewStatsdMonitor(statsdClient, config.StatsdPrefix, metricsSeparator)
	log.Infof("statsd metrics monitoring is enabled. (%s)", config.StatsdAddress)

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
