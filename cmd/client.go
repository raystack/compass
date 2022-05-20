package cmd

import (
	"context"
	"time"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func createConnection(ctx context.Context, host string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	return grpc.DialContext(ctx, host, opts...)
}

func createClient(ctx context.Context, host string) (compassv1beta1.CompassServiceClient, func(), error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(ctx, time.Second*2)
	conn, err := createConnection(dialTimeoutCtx, host)
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
