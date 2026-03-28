package middleware

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/internal/client"
)

// NamespaceKey is injected in context with the tenant context
type NamespaceKey struct{}

//go:generate mockery --name=NamespaceService -r --case underscore --with-expecter --structname NamespaceService --filename namespace_service.go --output=./mocks
type NamespaceService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*namespace.Namespace, error)
	GetByName(ctx context.Context, name string) (*namespace.Namespace, error)
}

// Namespace returns a new unary interceptor that extracts namespace from:
// 1. JWT token (priority 1) - uses namespaceClaimKey
// 2. x-namespace header (priority 2)
// 3. Defaults to DefaultNamespace
func Namespace(service NamespaceService, namespaceClaimKey, userUUIDHeaderKey string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			namespaceForRequest := namespace.DefaultNamespace

			// Extract if jwt is set with namespace id
			// jwt token has higher priority than header
			if authorization := req.Header().Get("Authorization"); authorization != "" {
				rawToken := strings.TrimPrefix(authorization, "Bearer ")
				if rawToken != "" {
					token, err := jwt.ParseInsecure([]byte(rawToken))
					if err == nil {
						// Check if namespace is passed as claim
						if namespaceID, okClaim := token.Get(namespaceClaimKey); okClaim {
							ns, err := getNamespaceByNameOrID(ctx, service, namespaceID.(string))
							if err != nil {
								return nil, err
							}
							namespaceForRequest = ns
						}

						// Override the user uuid if passed as claim
						if userUUID := token.Subject(); userUUID != "" && userUUIDHeaderKey != "" {
							req.Header().Set(userUUIDHeaderKey, userUUID)
						}
					}
				}
			}

			// If we have not already found a namespace, check in header
			if namespaceForRequest.ID == namespace.DefaultNamespace.ID {
				// Check if namespace is passed as header
				namespaceHeader := req.Header().Get(client.NamespaceHeaderKey)
				if namespaceHeader != "" {
					ns, err := getNamespaceByNameOrID(ctx, service, strings.TrimSpace(namespaceHeader))
					if err != nil {
						return nil, err
					}
					namespaceForRequest = ns
				}
			}

			// Build context with namespace
			ctx = BuildContextWithNamespace(ctx, namespaceForRequest)
			return next(ctx, req)
		}
	}
}

func getNamespaceByNameOrID(ctx context.Context, service NamespaceService, urn string) (*namespace.Namespace, error) {
	var ns *namespace.Namespace
	var err error
	nsID, parseErr := uuid.Parse(urn)
	if parseErr != nil {
		// If fail to parse a valid uuid, must be a name
		ns, err = service.GetByName(ctx, urn)
	} else {
		ns, err = service.GetByID(ctx, nsID)
	}
	if err != nil {
		if errors.Is(err, namespace.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return ns, nil
}

// BuildContextWithNamespace stores the namespace in context.
func BuildContextWithNamespace(ctx context.Context, ns *namespace.Namespace) context.Context {
	return context.WithValue(ctx, NamespaceKey{}, ns)
}

// FetchNamespaceFromContext retrieves namespace from context, returns default if not found.
func FetchNamespaceFromContext(ctx context.Context) *namespace.Namespace {
	if ns, ok := ctx.Value(NamespaceKey{}).(*namespace.Namespace); ok && ns != nil {
		return ns
	}
	return namespace.DefaultNamespace
}
