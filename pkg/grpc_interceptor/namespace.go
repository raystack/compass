package grpc_interceptor

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/internal/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NamespaceKey is injected in context with the tenant context
type NamespaceKey struct{}

//go:generate mockery --name=NamespaceService -r --case underscore --with-expecter --structname NamespaceService --filename namespace_service.go --output=./mocks
type NamespaceService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*namespace.Namespace, error)
	GetByName(ctx context.Context, name string) (*namespace.Namespace, error)
}

// NamespaceUnaryInterceptor namespace can be passed in jwt token or headers, if none provided
// it falls back to default
func NamespaceUnaryInterceptor(service NamespaceService, namespaceClaimKey, userUUIDHeaderKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		namespaceForRequest := namespace.DefaultNamespace
		if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
			// extract if jwt is set with namespace id
			// jwt token has higher priority then header

			if bearers := incomingMD.Get("Authorization"); len(bearers) > 0 {
				// Parse the token
				rawToken := strings.TrimPrefix(bearers[0], "Bearer ")
				if rawToken != "" {
					token, err := jwt.ParseInsecure([]byte(rawToken))
					if err == nil {
						// check if namespace is passed as claim
						if namespaceID, okClaim := token.Get(namespaceClaimKey); okClaim {
							ns, err := getNamespaceByNameOrID(ctx, service, namespaceID.(string))
							if err != nil {
								return nil, err
							}
							namespaceForRequest = ns
						}

						// override the user uuid if passed as claim
						if userUUID := token.Subject(); userUUID != "" && userUUIDHeaderKey != "" {
							incomingMD.Set(userUUIDHeaderKey, userUUID)
							ctx = metadata.NewIncomingContext(ctx, incomingMD)
						}
					}
				}
			}

			// if we have not already found a namespace, check in header
			if namespaceForRequest.ID == namespace.DefaultNamespace.ID {
				// check if namespace is passed as header
				namespaceHeaders := incomingMD.Get(client.NamespaceHeaderKey)
				if len(namespaceHeaders) > 0 {
					ns, err := getNamespaceByNameOrID(ctx, service, strings.TrimSpace(namespaceHeaders[0]))
					if err != nil {
						return nil, err
					}
					namespaceForRequest = ns
				}
			}
		}

		// fallback to default namespace
		ctx = BuildContextWithNamespace(ctx, namespaceForRequest)
		return handler(ctx, req)
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
