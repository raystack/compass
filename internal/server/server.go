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
	"github.com/odpf/compass/internal/server/health"
	handlersv1beta1 "github.com/odpf/compass/internal/server/v1beta1"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/compass/pkg/grpc_interceptor"
	"github.com/odpf/compass/pkg/statsd"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/log"
	"github.com/odpf/salt/mux"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

type Config struct {
	Host     string `mapstructure:"host" default:"0.0.0.0"`
	Port     int    `mapstructure:"port" default:"8080"`
	GRPCPort int    `mapstructure:"grpc_port" default:"8081"`
	BaseUrl  string `mapstructure:"baseurl" default:"localhost:8080"`

	// User Identity
	Identity IdentityConfig `mapstructure:"identity"`
}

func (cfg Config) addr() string     { return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port) }
func (cfg Config) grpcAddr() string { return fmt.Sprintf("%s:%d", cfg.Host, cfg.GRPCPort) }

type IdentityConfig struct {
	// User Identity
	HeaderKeyUUID       string `mapstructure:"headerkey_uuid" default:"Compass-User-UUID"`
	HeaderValueUUID     string `mapstructure:"headervalue_uuid" default:"odpf@email.com"`
	HeaderKeyEmail      string `mapstructure:"headerkey_email" default:"Compass-User-Email"`
	ProviderDefaultName string `mapstructure:"provider_default_name" default:""`
}

func Serve(
	ctx context.Context,
	config Config,
	logger *log.Logrus,
	pgClient *postgres.Client,
	nrApp *newrelic.Application,
	statsdReporter *statsd.Reporter,
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

	healthHandler := health.NewHandler()

	// init grpc
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logger.Entry()),
			nrgrpc.UnaryServerInterceptor(nrApp),
			grpc_interceptor.StatsD(statsdReporter),
			grpc_interceptor.UserHeaderCtx(config.Identity.HeaderKeyUUID, config.Identity.HeaderKeyEmail),
		)),
	)
	reflection.Register(grpcServer)

	compassv1beta1.RegisterCompassServiceServer(grpcServer, v1beta1Handler)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthHandler)

	// init http proxy
	grpcDialCtx, grpcDialCancel := context.WithTimeout(ctx, time.Second*5)
	defer grpcDialCancel()

	headerMatcher := makeHeaderMatcher(config)

	grpcConn, err := grpc.DialContext(grpcDialCtx, config.grpcAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	runtimeCtx, runtimeCancel := context.WithCancel(ctx)
	defer runtimeCancel()

	gwmux := runtime.NewServeMux(
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
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

	httpMux := http.NewServeMux()
	httpMux.Handle("/", gwmux)

	idleConnsClosed := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	go func() {
		signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-interrupt

		if pgClient != nil {
			logger.Warn("closing db...")
			if err := pgClient.Close(); err != nil {
				logger.Error("error when closing db", "err", err)
			}
			logger.Warn("db closed...")
		}

		close(idleConnsClosed)
	}()

	go func() {
		defer func() { interrupt <- syscall.SIGTERM }()

		if err := mux.Serve(
			ctx,
			mux.WithHTTPTarget(config.addr(), &http.Server{
				Handler:      grpcHandlerFunc(grpcServer, httpMux),
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
			}),
			mux.WithGRPCTarget(config.grpcAddr(), grpcServer),
			mux.WithGracePeriod(5*time.Second),
		); err != ctx.Err() {
			logger.Error("mux serve error", "err", err)
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
		case strings.ToLower(c.Identity.HeaderKeyUUID):
			return key, true
		case strings.ToLower(c.Identity.HeaderKeyEmail):
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}
}

// grpcHandlerFunc routes http1 calls to httpMux and http2 with grpc header to grpcServer.
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
