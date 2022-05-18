package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/internal/server/health"
	handlersv1beta1 "github.com/odpf/compass/internal/server/v1beta1"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/compass/pkg/grpc_interceptor"
	"github.com/odpf/compass/pkg/metrics"
	"github.com/odpf/salt/log"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type Config struct {
	Host string `mapstructure:"SERVER_HOST" default:"0.0.0.0"`
	Port int    `mapstructure:"SERVER_PORT" default:"8080"`

	// User Identity
	IdentityHeaderUUIDKey       string `mapstructure:"IDENTITY_HEADER_UUID" default:"Compass-User-UUID"`
	IdentityHeaderEmailKey      string `mapstructure:"IDENTITY_HEADER_EMAIL" default:"Compass-User-Email"`
	IdentityProviderDefaultName string `mapstructure:"IDENTITY_PROVIDER_DEFAULT_NAME" default:""`
}

func (cfg Config) addr() string { return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port) }

func Serve(
	ctx context.Context,
	config Config,
	logger log.Logger,
	pgClient *postgres.Client,
	nr *newrelic.Application,
	statsd *metrics.StatsdMonitor,
	assetService handlersv1beta1.AssetService,
	starService handlersv1beta1.StarService,
	discussionService handlersv1beta1.DiscussionService,
	tagService handlersv1beta1.TagService,
	tagTemplateService handlersv1beta1.TagTemplateService,
	userService handlersv1beta1.UserService,
) error {

	v1beta1Handler := handlersv1beta1.NewAPIServer(
		logger,
		assetService,
		starService,
		discussionService,
		tagService,
		tagTemplateService,
		userService,
	)

	healthHandler := &health.HealthHandler{}

	// init grpc
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.New())), //TODO: expose *logrus.Logger in salt
			nrgrpc.UnaryServerInterceptor(nr),
			grpc_interceptor.StatsD(statsd),
			grpc_interceptor.ValidateUser(config.IdentityHeaderUUIDKey, config.IdentityHeaderEmailKey, userService),
		)),
	)

	compassv1beta1.RegisterCompassServiceServer(grpcServer, v1beta1Handler)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthHandler)

	// init http proxy
	grpcDialCtx, grpcDialCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer grpcDialCancel()

	headerMatcher := makeHeaderMatcher(config)

	address := config.addr()
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

func makeHeaderMatcher(c Config) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		switch strings.ToLower(key) {
		case strings.ToLower(c.IdentityHeaderUUIDKey):
			return key, true
		case strings.ToLower(c.IdentityHeaderEmailKey):
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
