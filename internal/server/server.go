package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/raystack/compass/internal/client"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/raystack/compass/internal/server/health"
	handlersv1beta1 "github.com/raystack/compass/internal/server/v1beta1"
	"github.com/raystack/compass/internal/store/postgres"
	"github.com/raystack/compass/pkg/grpc_interceptor"
	"github.com/raystack/compass/pkg/statsd"
	compassv1beta1 "github.com/raystack/compass/proto/raystack/compass/v1beta1"
	"github.com/raystack/salt/log"
	"github.com/raystack/salt/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip" // Install the gzip compressor
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

type Config struct {
	Host    string `mapstructure:"host" default:"0.0.0.0"`
	Port    int    `mapstructure:"port" default:"8080"`
	BaseUrl string `mapstructure:"baseurl" default:"localhost:8080"`

	// User Identity
	Identity IdentityConfig `mapstructure:"identity"`

	// GRPC Config
	GRPC GRPCConfig `mapstructure:"grpc"`
}

func (cfg Config) addr() string     { return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port) }
func (cfg Config) grpcAddr() string { return fmt.Sprintf("%s:%d", cfg.Host, cfg.GRPC.Port) }

type IdentityConfig struct {
	// User Identity
	HeaderKeyUserUUID   string `yaml:"headerkey_uuid" mapstructure:"headerkey_uuid" default:"Compass-User-UUID"`
	HeaderValueUserUUID string `yaml:"headervalue_uuid" mapstructure:"headervalue_uuid" default:"raystack@email.com"`
	HeaderKeyUserEmail  string `yaml:"headerkey_email" mapstructure:"headerkey_email" default:"Compass-User-Email"`
	ProviderDefaultName string `yaml:"provider_default_name" mapstructure:"provider_default_name" default:""`

	NamespaceClaimKey string `yaml:"namespace_claim_key" mapstructure:"namespace_claim_key" default:"namespace_id"`
}

type GRPCConfig struct {
	Port           int `yaml:"port" mapstructure:"port" default:"8081"`
	MaxRecvMsgSize int `yaml:"max_recv_msg_size" mapstructure:"max_recv_msg_size" default:"33554432"`
	MaxSendMsgSize int `yaml:"max_send_msg_size" mapstructure:"max_send_msg_size" default:"33554432"`
}

func Serve(
	ctx context.Context,
	config Config,
	logger *log.Logrus,
	pgClient *postgres.Client,
	nrApp *newrelic.Application,
	statsdReporter *statsd.Reporter,
	namespaceService handlersv1beta1.NamespaceService,
	assetService handlersv1beta1.AssetService,
	starService handlersv1beta1.StarService,
	discussionService handlersv1beta1.DiscussionService,
	tagService handlersv1beta1.TagService,
	tagTemplateService handlersv1beta1.TagTemplateService,
	userService handlersv1beta1.UserService,
) error {
	v1beta1Handler := handlersv1beta1.NewAPIServer(
		logger,
		namespaceService,
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
		grpc.MaxRecvMsgSize(config.GRPC.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.GRPC.MaxSendMsgSize),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logger.Entry()),
			nrgrpc.UnaryServerInterceptor(nrApp),
			grpc_interceptor.StatsD(statsdReporter),
			grpc_interceptor.NamespaceUnaryInterceptor(namespaceService, config.Identity.NamespaceClaimKey, config.Identity.HeaderKeyUserUUID),
			grpc_interceptor.UserHeaderCtx(config.Identity.HeaderKeyUserUUID, config.Identity.HeaderKeyUserEmail),
		)),
	)
	reflection.Register(grpcServer)

	compassv1beta1.RegisterCompassServiceServer(grpcServer, v1beta1Handler)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthHandler)

	// init http proxy
	grpcDialCtx, grpcDialCancel := context.WithTimeout(ctx, time.Second*5)
	defer grpcDialCancel()

	headerMatcher := makeHeaderMatcher(config)

	grpcConn, err := grpc.DialContext(
		grpcDialCtx,
		config.grpcAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(config.GRPC.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(config.GRPC.MaxSendMsgSize),
		))
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

	defer func() {
		if pgClient != nil {
			logger.Warn("closing db...")
			if err := pgClient.Close(); err != nil {
				logger.Error("error when closing db", "err", err)
			}
			logger.Warn("db closed...")
		}
	}()

	logger.Info("Starting server", "http_port", config.addr(), "grpc_port", config.grpcAddr())
	if err := mux.Serve(
		ctx,
		mux.WithHTTPTarget(config.addr(), &http.Server{
			Handler:      handlers.CompressHandler(gwmux),
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
		}),
		mux.WithGRPCTarget(config.grpcAddr(), grpcServer),
		mux.WithGracePeriod(5*time.Second),
	); !errors.Is(err, context.Canceled) {
		logger.Error("mux serve error", "err", err)
	}

	logger.Info("server stopped")
	return nil
}

// makeHeaderMatcher overrides the default grpc gateway behaviour of only mapping http headers
// that start with "grpc-metadata-" prefix
func makeHeaderMatcher(c Config) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		switch strings.ToLower(key) {
		case strings.ToLower(c.Identity.HeaderKeyUserUUID):
			return key, true
		case strings.ToLower(c.Identity.HeaderKeyUserEmail):
			return key, true
		case client.NamespaceHeaderKey:
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}
}
