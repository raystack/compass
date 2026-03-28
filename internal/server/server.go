package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/raystack/compass/internal/config"
	"github.com/rs/cors"
	"github.com/raystack/compass/handler"
	"github.com/raystack/compass/internal/middleware"
	"github.com/raystack/compass/gen/raystack/compass/v1beta1/compassv1beta1connect"
	log "github.com/raystack/salt/observability/logger"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func Serve(
	ctx context.Context,
	cfg config.ServerConfig,
	logger *log.Logrus,
	namespaceService handler.NamespaceService,
	assetService handler.AssetService,
	starService handler.StarService,
	discussionService handler.DiscussionService,
	tagService handler.TagService,
	tagTemplateService handler.TagTemplateService,
	userService handler.UserService,
) error {
	v1beta1Handler := handler.New(
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

	validateInterceptor := validate.NewInterceptor()

	interceptors := connect.WithInterceptors(
		otelInterceptor,
		middleware.Recovery(),
		middleware.Logger(logger),
		validateInterceptor,
		middleware.ErrorResponse(logger),
		middleware.Namespace(namespaceService, cfg.Identity.NamespaceClaimKey, cfg.Identity.HeaderKeyUserUUID),
		middleware.UserHeaderCtx(cfg.Identity.HeaderKeyUserUUID, cfg.Identity.HeaderKeyUserEmail),
	)

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register Connect service handler
	path, svcHandler := compassv1beta1connect.NewCompassServiceHandler(
		v1beta1Handler,
		interceptors,
		connect.WithReadMaxBytes(cfg.MaxRecvMsgSize),
		connect.WithSendMaxBytes(cfg.MaxSendMsgSize),
	)
	mux.Handle(path, svcHandler)

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

	// CORS middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   connectcors.AllowedMethods(),
		AllowedHeaders:   connectcors.AllowedHeaders(),
		ExposedHeaders:   connectcors.ExposedHeaders(),
		AllowCredentials: true,
	})

	// Create HTTP server with h2c support for HTTP/2 without TLS
	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      h2c.NewHandler(corsHandler.Handler(mux), &http2.Server{}),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Info("Starting server", "addr", cfg.Addr())
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
