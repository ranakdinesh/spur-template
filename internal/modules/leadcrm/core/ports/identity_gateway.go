package ports

import "context"

type IdentityGateway interface {
	VerifySession(ctx context.Context, token string) (string, error)
	ValidateAPIKey(ctx context.Context, key string) (string, error) // Returns TenantID
}
