package client

import (
	"context"
	"time"

	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Config struct {
	Host                  string `mapstructure:"host" default:"localhost:8081"`
	ServerHeaderKeyUUID   string `mapstructure:"serverheaderkey_uuid" default:"Compass-User-UUID"`
	ServerHeaderValueUUID string `mapstructure:"serverheadervalue_uuid" default:"compass@odpf.com"`
}

var config Config

func SetConfig(c Config) {
	config = c
}

func Create(ctx context.Context) (compassv1beta1.CompassServiceClient, func(), error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(ctx, time.Second*2)
	conn, err := createConnection(dialTimeoutCtx)
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

func SetMetadata(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{config.ServerHeaderKeyUUID: config.ServerHeaderValueUUID})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}

func createConnection(ctx context.Context) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	return grpc.DialContext(ctx, config.Host, opts...)
}
