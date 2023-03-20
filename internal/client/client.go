package client

import (
	"context"
	"time"

	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// NamespaceHeaderKey specify what namespace request is targeted for
// if not provided, default namespace is assumed
const NamespaceHeaderKey = "x-namespace-id"

type Config struct {
	Host                      string `mapstructure:"host" default:"localhost:8081"`
	ServerHeaderKeyUserUUID   string `yaml:"serverheaderkey_uuid" mapstructure:"serverheaderkey_uuid" default:"Compass-User-UUID"`
	ServerHeaderValueUserUUID string `yaml:"serverheadervalue_uuid" mapstructure:"serverheadervalue_uuid" default:"compass@odpf.com"`
}

func Create(ctx context.Context, cfg Config) (compassv1beta1.CompassServiceClient, func(), error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(ctx, time.Second*2)
	conn, err := createConnection(dialTimeoutCtx, cfg)
	if err != nil {
		dialCancel()
		return nil, nil, err
	}

	cancel := func() {
		dialCancel()
		conn.Close()
	}

	client := compassv1beta1.NewCompassServiceClient(conn)
	return client, cancel, nil
}

func SetMetadata(ctx context.Context, cfg Config, namespaceID string) context.Context {
	md := metadata.New(map[string]string{
		cfg.ServerHeaderKeyUserUUID: cfg.ServerHeaderValueUserUUID,
		NamespaceHeaderKey:          namespaceID,
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func createConnection(ctx context.Context, cfg Config) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	return grpc.DialContext(ctx, cfg.Host, opts...)
}
