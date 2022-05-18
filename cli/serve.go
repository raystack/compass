package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/elastic/go-elasticsearch/v7"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/newrelic/go-agent/v3/integrations/nrelasticsearch-v7"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/compass/api"
	"github.com/odpf/compass/api/grpc_interceptor"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/metrics"
	esStore "github.com/odpf/compass/store/elasticsearch"
	"github.com/odpf/compass/store/postgres"
	"github.com/odpf/compass/tag"
	"github.com/odpf/compass/user"
	"github.com/odpf/salt/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
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

	logger := initLogger(config.LogLevel)
	logger.Info("compass starting", "version", Version)

	esClient := initElasticsearch(config, logger)
	newRelicMonitor := initNewRelicMonitor(config, logger)
	statsdMonitor := initStatsdMonitor(config, logger)
	pgClient := initPostgres(logger, config)
	deps := initDependencies(logger, config, esClient, pgClient, newRelicMonitor.Application(), statsdMonitor)

	handlers := api.NewHandlers(logger, deps)

	// init grpc
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.New())), //TODO: expose *logrus.Logger in salt
			nrgrpc.UnaryServerInterceptor(newRelicMonitor.Application()),
			grpc_interceptor.StatsD(statsdMonitor),
			grpc_interceptor.ValidateUser(config.IdentityUUIDHeaderKey, config.IdentityEmailHeaderKey, deps.UserService),
		)),
	)

	compassv1beta1.RegisterCompassServiceServer(grpcServer, handlers.GRPCHandler)
	grpc_health_v1.RegisterHealthServer(grpcServer, handlers.HealthHandler)

	// init http proxy
	grpcDialCtx, grpcDialCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer grpcDialCancel()

	headerMatcher := makeHeaderMatcher(config)

	address := fmt.Sprintf(":%s", config.ServerPort)
	grpcConn, err := grpc.DialContext(grpcDialCtx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	runtimeCtx, runtimeCancel := context.WithCancel(context.Background())
	defer runtimeCancel()

	gwmux := runtime.NewServeMux(
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
		runtime.WithHealthEndpointAt(grpc_health_v1.NewHealthClient(grpcConn), "/ping"),
	)

	if err := compassv1beta1.RegisterCompassServiceHandler(runtimeCtx, gwmux, grpcConn); err != nil {
		return err
	}

	baseMux := http.NewServeMux()
	baseMux.Handle("/", gwmux)

	httpServer := &http.Server{
		Handler:      grpcHandlerFunc(grpcServer, baseMux),
		Addr:         address,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	idleConnsClosed := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	go func() {
		signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-interrupt

		if httpServer != nil {
			// We received an interrupt signal, shut down.
			logger.Warn("stopping http server...")
			if err := httpServer.Shutdown(context.Background()); err != nil {
				logger.Error("HTTP server Shutdown", "err", err)
			}
		}

		if grpcServer != nil {
			logger.Warn("stopping grpc server...")
			grpcServer.GracefulStop()
		}

		if pgClient != nil {
			logger.Warn("closing db...")
			if err := pgClient.Close(); err != nil {
				logger.Error("error when closing db", "err", err)
			}
		}

		close(idleConnsClosed)
	}()

	go func() {
		defer func() { interrupt <- syscall.SIGTERM }()
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("HTTP server ListenAndServe", "err", err)
		}
	}()

	logger.Info("server started")

	<-idleConnsClosed

	logger.Info("server stopped")

	return nil
}

func initDependencies(
	logger log.Logger,
	config Config,
	esClient *elasticsearch.Client,
	pgClient *postgres.Client,
	nrApp *newrelic.Application,
	statsdMonitor *metrics.StatsdMonitor,
) *api.Dependencies {
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
	userService := user.NewService(logger, userRepository)

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
		Logger:               logger,
		NRApp:                nrApp,
		StatsdMonitor:        statsdMonitor,
		AssetRepository:      assetRepository,
		DiscoveryRepository:  discoveryRepo,
		LineageRepository:    lineageRepo,
		TagService:           tagService,
		TagTemplateService:   tagTemplateService,
		UserService:          userService,
		StarRepository:       starRepository,
		DiscussionRepository: discussionRepository,
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

func makeHeaderMatcher(c Config) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		switch strings.ToLower(key) {
		case strings.ToLower(c.IdentityUUIDHeaderKey):
			return key, true
		case strings.ToLower(c.IdentityEmailHeaderKey):
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}
}

// grpcHandlerFunc routes http1 calls to baseMux and http2 with grpc header to grpcServer.
// Using a single port for proxying both http1 & 2 protocols will degrade http performance
// but for our usecase the convenience per performance tradeoff is better suited
// if in future, this does become a bottleneck(which I highly doubt), we can break the service
// into two ports, default port for grpc and default+1 for grpc-gateway proxy.
// We can also use something like a connection multiplexer
// https://github.com/soheilhy/cmux to achieve the same.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}
