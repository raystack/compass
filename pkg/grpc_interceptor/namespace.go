package grpc_interceptor

//go:generate mockery --name=NamespaceService -r --case underscore --with-expecter --structname NamespaceService --filename namespace_service.go --output=./mocks

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/odpf/compass/core/namespace"
	"github.com/odpf/compass/internal/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NamespaceKey is injected in context with the tenant context
type NamespaceKey struct{}

type NamespaceService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*namespace.Namespace, error)
	GetByName(ctx context.Context, name string) (*namespace.Namespace, error)
}

// NamespaceUnaryInterceptor namespace can be passed in jwt token or headers, if none provided
// it falls back to default
func NamespaceUnaryInterceptor(service NamespaceService) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			// extract if jwt is set with namespace id
			// jwt token has higher priority then header
			bearers := md.Get("Authorization")
			if len(bearers) > 0 {
				// Parse the token
				rawToken := strings.TrimPrefix(bearers[0], "Bearer ")
				token, err := jwt.ParseSigned(rawToken)
				if err == nil {
					claims := make(map[string]string)
					_ = token.UnsafeClaimsWithoutVerification(&claims)
					if namespaceID, ok1 := claims["namespace_id"]; ok1 {
						ns, err := getNamespaceByNameOrID(ctx, service, namespaceID)
						if err != nil {
							return nil, err
						}
						return handler(BuildContextWithNamespace(ctx, ns), req)
					}
				}
			}

			// check if namespace is passed as header
			namespaceHeaders := md.Get(client.NamespaceHeaderKey)
			if len(namespaceHeaders) > 0 {
				ns, err := getNamespaceByNameOrID(ctx, service, strings.TrimSpace(namespaceHeaders[0]))
				if err != nil {
					return nil, err
				}
				return handler(BuildContextWithNamespace(ctx, ns), req)
			}
		}

		// fallback to default namespace
		return handler(BuildContextWithNamespace(ctx, namespace.DefaultNamespace), req)
	}
}

func getNamespaceByNameOrID(ctx context.Context, service NamespaceService, urn string) (*namespace.Namespace, error) {
	var ns *namespace.Namespace
	nsID, err := uuid.Parse(urn)
	if err != nil {
		// if fail to parse a valid uuid, must be a name
		if ns, err = service.GetByName(ctx, urn); err != nil {
			return nil, status.Error(codes.NotFound, err.Error())
		}
	} else {
		if ns, err = service.GetByID(ctx, nsID); err != nil {
			return nil, status.Error(codes.NotFound, err.Error())
		}
	}
	return ns, nil
}

func BuildContextWithNamespace(ctx context.Context, ns *namespace.Namespace) context.Context {
	return context.WithValue(ctx, NamespaceKey{}, ns)
}

// FetchNamespaceFromContext if not found, fallback to default
func FetchNamespaceFromContext(ctx context.Context) *namespace.Namespace {
	if ns, ok := ctx.Value(NamespaceKey{}).(*namespace.Namespace); ok && ns != nil {
		return ns
	}
	return namespace.DefaultNamespace
}
