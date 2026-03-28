package client

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/compass/proto/compassv1beta1/compassv1beta1connect"
)

// NamespaceHeaderKey specify what namespace request is targeted for
// if not provided, default namespace is assumed
const NamespaceHeaderKey = "x-namespace"

type Config struct {
	Host                      string `mapstructure:"host" default:"localhost:8080"`
	ServerHeaderKeyUserUUID   string `yaml:"serverheaderkey_uuid" mapstructure:"serverheaderkey_uuid" default:"Compass-User-UUID"`
	ServerHeaderValueUserUUID string `yaml:"serverheadervalue_uuid" mapstructure:"serverheadervalue_uuid" default:"compass@raystack.com"`
}

// Client wraps the Connect client with header configuration
type Client struct {
	compassv1beta1connect.CompassServiceClient
	cfg Config
}

// Create creates a new Connect client for the Compass service.
func Create(ctx context.Context, cfg Config) (*Client, error) {
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	// Build base URL
	baseURL := "http://" + cfg.Host

	connectClient := compassv1beta1connect.NewCompassServiceClient(
		httpClient,
		baseURL,
		connect.WithGRPC(), // Use gRPC protocol for compatibility
	)

	return &Client{
		CompassServiceClient: connectClient,
		cfg:                  cfg,
	}, nil
}

// NewRequest creates a new Connect request with the configured headers.
func NewRequest[T any](cfg Config, namespaceID string, msg *T) *connect.Request[T] {
	req := connect.NewRequest(msg)
	req.Header().Set(cfg.ServerHeaderKeyUserUUID, cfg.ServerHeaderValueUserUUID)
	if namespaceID != "" {
		req.Header().Set(NamespaceHeaderKey, namespaceID)
	}
	return req
}
