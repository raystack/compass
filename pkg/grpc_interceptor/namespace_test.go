package grpc_interceptor_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/internal/client"
	"github.com/raystack/compass/pkg/grpc_interceptor"
	"github.com/raystack/compass/pkg/grpc_interceptor/mocks"
	"google.golang.org/grpc/metadata"
	"reflect"
	"testing"
)

func TestNamespaceUnaryInterceptor(t *testing.T) {
	type args struct {
		service *mocks.NamespaceService
		ctx     context.Context
	}
	type wants struct {
		ns  *namespace.Namespace
		err error
	}
	mockedID := uuid.MustParse("e167aaea-ca00-4ec1-8f0d-04c067da54b1")
	ns := &namespace.Namespace{
		ID:       mockedID,
		Name:     "umbrella",
		State:    namespace.SharedState,
		Metadata: nil,
	}
	// jwtWithNamespaceID contains namespace_id: "e167aaea-ca00-4ec1-8f0d-04c067da54b1"
	jwtWithNamespaceID := `eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwibmFtZXNwYWNlX2lkIjoiZTE2N2FhZWEtY2EwMC00ZWMxLThmMGQtMDRjMDY3ZGE1NGIxIiwiaWF0IjoxNjc5NzI2MTcxLCJleHAiOjI2Nzk3Mjk3NzF9.2KVJnE9wICA7SqK6aIex0Bzx7Sz9csL-3rfib9jI2fQ`

	tests := []struct {
		name  string
		args  args
		want  wants
		mocks func(context.Context, *mocks.NamespaceService)
	}{
		{
			name: "fallback default namespace if none passed in headers",
			args: args{
				service: new(mocks.NamespaceService),
				ctx:     context.Background(),
			},
			want: wants{
				ns: namespace.DefaultNamespace,
			},
			mocks: func(context.Context, *mocks.NamespaceService) {},
		},
		{
			name: "extract from header if passed namespace id in headers",
			args: args{
				service: new(mocks.NamespaceService),
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					client.NamespaceHeaderKey: {ns.ID.String()},
				}),
			},
			want: wants{
				ns: ns,
			},
			mocks: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().GetByID(ctx, ns.ID).Return(ns, nil)
			},
		},
		{
			name: "extract from header if passed namespace name in headers",
			args: args{
				service: new(mocks.NamespaceService),
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					client.NamespaceHeaderKey: {ns.Name},
				}),
			},
			want: wants{
				ns: ns,
			},
			mocks: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().GetByName(ctx, ns.Name).Return(ns, nil)
			},
		},
		{
			name: "extract from jwt if passed namespace id in jwt",
			args: args{
				service: new(mocks.NamespaceService),
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					"Authorization": {"Bearer " + jwtWithNamespaceID},
				}),
			},
			want: wants{
				ns: ns,
			},
			mocks: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().GetByID(ctx, mockedID).Return(ns, nil)
			},
		},
		{
			name: "extract from jwt if passed namespace id in jwt and namespace header",
			args: args{
				service: new(mocks.NamespaceService),
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					"Authorization":           {"Bearer " + jwtWithNamespaceID},
					client.NamespaceHeaderKey: {namespace.DefaultNamespace.Name},
				}),
			},
			want: wants{
				ns: ns,
			},
			mocks: func(ctx context.Context, nss *mocks.NamespaceService) {
				nss.EXPECT().GetByID(ctx, mockedID).Return(ns, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mocks(tt.args.ctx, tt.args.service)

			interceptor := grpc_interceptor.NamespaceUnaryInterceptor(tt.args.service)
			_, err := interceptor(tt.args.ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				got := grpc_interceptor.FetchNamespaceFromContext(ctx)
				if !reflect.DeepEqual(got, tt.want.ns) {
					t.Errorf("NamespaceUnaryInterceptor() = %v, want %v", got, tt.want.ns)
				}
				return nil, nil
			})
			if err != tt.want.err {
				t.Errorf("NamespaceUnaryInterceptor() = %v, want %v", err, tt.want.err)
			}
			tt.args.service.AssertExpectations(t)
		})
	}
}
