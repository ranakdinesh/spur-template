package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
)

// CreateClient creates a new OAuth2 client
func (s *Store) CreateClient(ctx context.Context, arg sqlc.CreateClientParams) (sqlc.FositeClients, error) {
	return s.getQueries(ctx).CreateClient(ctx, arg)
}

// GetClient retrieves a client by ID
func (s *Store) GetClient(ctx context.Context, id string) (sqlc.FositeClients, error) {
	return s.getQueries(ctx).GetClient(ctx, id)
}

// GetActiveClient retrieves a client by ID only if it is active
func (s *Store) GetActiveClient(ctx context.Context, id string) (sqlc.FositeClients, error) {
	return s.getQueries(ctx).GetActiveClient(ctx, id)
}

// ListClients retrieves all clients for a given tenant
func (s *Store) ListClients(ctx context.Context, tenantID uuid.UUID) ([]sqlc.FositeClients, error) {
	return s.getQueries(ctx).ListClients(ctx, tenantID)
}

// ListPublicClients retrieves all public clients
func (s *Store) ListPublicClients(ctx context.Context) ([]sqlc.FositeClients, error) {
	return s.getQueries(ctx).ListPublicClients(ctx)
}

// UpdateClientSecret updates the client secret
func (s *Store) UpdateClientSecret(ctx context.Context, arg sqlc.UpdateClientSecretParams) error {
	return s.getQueries(ctx).UpdateClientSecret(ctx, arg)
}

// ToggleClientStatus activates or deactivates a client
func (s *Store) ToggleClientStatus(ctx context.Context, arg sqlc.ToggleClientStatusParams) error {
	return s.getQueries(ctx).ToggleClientStatus(ctx, arg)
}

// UpdateClientConfig updates client configuration (redirect URIs, scopes, grant types)
func (s *Store) UpdateClientConfig(ctx context.Context, arg sqlc.UpdateClientConfigParams) error {
	return s.getQueries(ctx).UpdateClientConfig(ctx, arg)
}

// DeleteClient deletes a client by ID
func (s *Store) DeleteClient(ctx context.Context, id string) error {
	return s.getQueries(ctx).DeleteClient(ctx, id)
}
