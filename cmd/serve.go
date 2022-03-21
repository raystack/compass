package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	nrelasticsearch "github.com/newrelic/go-agent/v3/integrations/nrelasticsearch-v7"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/columbus/api"
	"github.com/odpf/columbus/api/grpc_interceptor"
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/metrics"
	esStore "github.com/odpf/columbus/store/elasticsearch"
	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/tag"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"github.com/odpf/salt/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
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
	deps := initDependencies(logger, esClient)

	handlers := api.NewHandlers(logger, deps)

	// old http: to be removed
	router := mux.NewRouter()
	if newRelicMonitor != nil {
		newRelicMonitor.MonitorRouter(router)
	}
	if statsdMonitor != nil {
		statsdMonitor.MonitorRouter(router)
	}
	router.Use(requestLoggerMiddleware(
		logger.Writer(),
	))

	api.RegisterHTTPRoutes(api.Config{IdentityHeaderKey: config.IdentityHeader}, router, deps, handlers.HTTPHandler)
	// old http: to be removed

	// grpc
	ctx, cancelFunc := context.WithCancel(
		server.HandleSignals(context.Background()),
	)
	defer cancelFunc()

	muxServer, gw, err := newGRPCServer(
		config,
		newRelicMonitor.Application(),
		grpc_recovery.UnaryServerInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.New())), //TODO: expose *logrus.Logger in salt
		nrgrpc.UnaryServerInterceptor(newRelicMonitor.Application()),
		grpc_interceptor.ValidateUser(config.IdentityHeader, deps.UserService),
	)
	if err != nil {
		panic(err)
	}

	err = gw.RegisterHandler(ctx, compassv1beta1.RegisterCompassServiceHandlerFromEndpoint)
	if err != nil {
		panic(err)
	}

	muxServer.RegisterService(
		&compassv1beta1.CompassService_ServiceDesc,
		handlers.GRPCHandler,
	)

	muxServer.RegisterHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}))

	muxServer.SetGateway("/", gw)

	logger.Info("starting server", "host", config.ServerHost, "port", config.ServerPort)

	serverErrorChan := make(chan error)

	go func() {
		logger.Info(fmt.Sprintf("starting grpc gateway server on %s:%s", config.ServerHost, config.GRPCServerPort))
		serverErrorChan <- muxServer.Serve()
	}()

	go func() {
		serverAddr := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
		logger.Info(fmt.Sprintf("starting http server on %s", serverAddr))
		serverErrorChan <- http.ListenAndServe(serverAddr, router)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*30)
		defer shutdownCancel()
		muxServer.Shutdown(shutdownCtx)
	case serverError := <-serverErrorChan:
		panic(serverError)
	}
}

func initDependencies(
	logger log.Logger,
	esClient *elasticsearch.Client,
) *api.Dependencies {
	typeRepository := esStore.NewTypeRepository(esClient)
	recordRepositoryFactory := esStore.NewRecordRepositoryFactory(esClient)
	recordSearcher, err := esStore.NewSearcher(esStore.SearcherConfig{
		Client: esClient,
	})
	if err != nil {
		logger.Fatal("error creating searcher", "error", err)
	}

	pgClient := initPostgres(logger, config)

	// init tag
	tagRepository, err := postgres.NewTagRepository(pgClient)
	if err != nil {
		logger.Fatal("failed to create new tag repository", "error", err)
	}
	tagTemplateRepository, err := postgres.NewTagTemplateRepository(pgClient)
	if err != nil {
		logger.Fatal("failed to create new tag template repository", "error", err)
	}
	tagTemplateService := tag.NewTemplateService(tagTemplateRepository)
	tagService := tag.NewService(tagRepository, tagTemplateService)

	// init user
	userRepository, err := postgres.NewUserRepository(pgClient)
	if err != nil {
		logger.Fatal("failed to create new user repository", "error", err)
	}
	userService := user.NewService(userRepository, user.Config{
		IdentityProviderDefaultName: config.IdentityProviderDefaultName,
	})

	assetRepository, err := postgres.NewAssetRepository(pgClient, userRepository, 0, config.IdentityProviderDefaultName)
	if err != nil {
		logger.Fatal("failed to create new asset repository", "error", err)
	}

	// init discussion
	discussionRepository, err := postgres.NewDiscussionRepository(pgClient, 0)
	if err != nil {
		logger.Fatal("failed to create new discussion repository", "error", err)
	}

	// init star
	starRepository, err := postgres.NewStarRepository(pgClient)
	if err != nil {
		logger.Fatal("failed to create new star repository", "error", err)
	}

	discoveryRepo := esStore.NewDiscoveryRepository(esClient)
	lineageRepo, err := postgres.NewLineageRepository(pgClient)
	if err != nil {
		logger.Fatal("failed to create new lineage repository", "error", err)
	}

	return &api.Dependencies{
		Logger:                  logger,
		AssetRepository:         assetRepository,
		DiscoveryRepository:     discoveryRepo,
		TypeRepository:          typeRepository,
		DiscoveryService:        discovery.NewService(recordRepositoryFactory, recordSearcher),
		RecordRepositoryFactory: recordRepositoryFactory,
		LineageRepository:       lineageRepo,
		TagService:              tagService,
		TagTemplateService:      tagTemplateService,
		UserService:             userService,
		StarRepository:          starRepository,
		DiscussionRepository:    discussionRepository,
	}
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

func newGRPCServer(cfg Config, nrApp *newrelic.Application, middleware ...grpc.UnaryServerInterceptor) (*server.MuxServer, *server.GRPCGateway, error) {
	grpcPortInt, err := strconv.Atoi(config.GRPCServerPort)
	if err != nil {
		return nil, nil, err
	}

	headerMatcher := makeHeaderMatcher(cfg)
	muxServer, err := server.NewMux(
		server.Config{
			Port: grpcPortInt,
			Host: cfg.ServerHost,
		},

		server.WithMuxGRPCServerOptions(
			grpc.UnaryInterceptor(
				grpc_middleware.ChainUnaryServer(
					middleware...,
				),
			),
		),
	)
	// server.WithMuxHTTPServer(httpServer))
	if err != nil {
		return nil, nil, err
	}

	gw, err := server.NewGateway(cfg.ServerHost, grpcPortInt,
		server.WithGatewayMuxOptions(
			runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
			runtime.WithIncomingHeaderMatcher(headerMatcher),
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			}),
		))
	if err != nil {
		return nil, nil, err
	}

	// if err = gw.RegisterHandler(ctx, compassv1beta1.RegisterCompassServiceHandlerFromEndpoint); err != nil {
	// 	return nil, err
	// }

	// muxServer.RegisterService(
	// 	&compassv1beta1.CompassService_ServiceDesc,
	// 	grpcapi.NewService(logger, deps),
	// )

	// api.RegisterHandlers(ctx, muxServer, gw)
	return muxServer, gw, nil
}

// func (s *Server) Run() error {
// 	ctx, cancelFunc := context.WithCancel(
// 		server.HandleSignals(context.Background()),
// 	)
// 	defer cancelFunc()
// 	if err = s.gw.RegisterHandler(ctx, compassv1beta1.RegisterCompassServiceHandlerFromEndpoint); err != nil {
// 		return nil, err
// 	}

// 	s.muxServer.RegisterService(
// 		&compassv1beta1.CompassService_ServiceDesc,
// 		grpcapi.NewService(logger, deps),
// 	)

// 	// api.RegisterHandlers(ctx, muxServer, gw)
// }

func makeHeaderMatcher(c Config) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		switch strings.ToLower(key) {
		case strings.ToLower(c.IdentityHeader):
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}
}
