package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"github.com/gorilla/handlers"
	"github.com/newrelic/go-agent/v3/newrelic"
	handlersv1beta1 "github.com/raystack/compass/internal/server/v1beta1"
	"github.com/raystack/compass/internal/store/postgres"
	"github.com/raystack/compass/pkg/server/interceptor"
	"github.com/raystack/compass/proto/compassv1beta1/compassv1beta1connect"
	log "github.com/raystack/salt/observability/logger"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Config struct {
	Host    string `mapstructure:"host" default:"0.0.0.0"`
	Port    int    `mapstructure:"port" default:"8080"`
	BaseUrl string `mapstructure:"baseurl" default:"localhost:8080"`

	// User Identity
	Identity IdentityConfig `mapstructure:"identity"`

	// Message size limits (for compatibility)
	MaxRecvMsgSize int `yaml:"max_recv_msg_size" mapstructure:"max_recv_msg_size" default:"33554432"`
	MaxSendMsgSize int `yaml:"max_send_msg_size" mapstructure:"max_send_msg_size" default:"33554432"`
}

func (cfg Config) addr() string { return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port) }

type IdentityConfig struct {
	// User Identity
	HeaderKeyUserUUID   string `yaml:"headerkey_uuid" mapstructure:"headerkey_uuid" default:"Compass-User-UUID"`
	HeaderValueUserUUID string `yaml:"headervalue_uuid" mapstructure:"headervalue_uuid" default:"raystack@email.com"`
	HeaderKeyUserEmail  string `yaml:"headerkey_email" mapstructure:"headerkey_email" default:"Compass-User-Email"`
	ProviderDefaultName string `yaml:"provider_default_name" mapstructure:"provider_default_name" default:""`

	NamespaceClaimKey string `yaml:"namespace_claim_key" mapstructure:"namespace_claim_key" default:"namespace_id"`
}

func Serve(
	ctx context.Context,
	config Config,
	logger *log.Logrus,
	pgClient *postgres.Client,
	nrApp *newrelic.Application,
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

	// Build interceptor chain
	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		return fmt.Errorf("failed to create otel interceptor: %w", err)
	}

	interceptors := connect.WithInterceptors(
		otelInterceptor,
		interceptor.Recovery(),
		interceptor.Logger(logger),
		interceptor.ErrorResponse(logger),
		interceptor.Namespace(namespaceService, config.Identity.NamespaceClaimKey, config.Identity.HeaderKeyUserUUID),
		interceptor.UserHeaderCtx(config.Identity.HeaderKeyUserUUID, config.Identity.HeaderKeyUserEmail),
	)

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register Connect service handler
	path, handler := compassv1beta1connect.NewCompassServiceHandler(
		v1beta1Handler,
		interceptors,
		connect.WithReadMaxBytes(config.MaxRecvMsgSize),
		connect.WithSendMaxBytes(config.MaxSendMsgSize),
	)
	mux.Handle(path, handler)

	// Register gRPC reflection for tooling compatibility (grpcurl, etc.)
	reflector := grpcreflect.NewStaticReflector(
		"raystack.compass.v1beta1.CompassService",
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Health check endpoint
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	// Create HTTP server with h2c support for HTTP/2 without TLS
	server := &http.Server{
		Addr:         config.addr(),
		Handler:      h2c.NewHandler(handlers.CompressHandler(mux), &http2.Server{}),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Cleanup on shutdown
	defer func() {
		if pgClient != nil {
			logger.Warn("closing db...")
			if err := pgClient.Close(); err != nil {
				logger.Error("error when closing db", "err", err)
			}
			logger.Warn("db closed...")
		}
	}()

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Info("Starting server", "addr", config.addr())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		logger.Info("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", "err", err)
		}
	case err := <-errChan:
		logger.Error("server error", "err", err)
		return err
	}

	logger.Info("server stopped")
	return nil
}
