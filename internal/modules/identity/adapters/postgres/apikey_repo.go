package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
)

type APIKeyRepo struct {
	store *Store
}

func NewAPIKeyRepo(store *Store) ports.APIKeyRepo {
	return &APIKeyRepo{store: store}
}

func (r *APIKeyRepo) CreateAPIKey(ctx context.Context, key *domain.APIKey) error {
	q := r.store.getQueries(ctx)

	_, err := q.CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{
		ID:             key.ID,
		TenantID:       key.TenantID,
		Name:           key.Name,
		Type:           key.Type,
		Prefix:         key.Prefix,
		KeyHash:        key.KeyHash,
		Scopes:         key.Scopes,
		AllowedOrigins: key.AllowedOrigins,
		ExpiresAt:      key.ExpiresAt,
	})
	return err
}

func (r *APIKeyRepo) GetAPIKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.APIKey, error) {
	q := r.store.getQueries(ctx)

	row, err := q.GetAPIKey(ctx, sqlc.GetAPIKeyParams{
		ID:       id,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return mapAPIKeyToDomain(row), nil
}

func (r *APIKeyRepo) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*domain.APIKey, error) {
	q := r.store.getQueries(ctx)

	row, err := q.GetAPIKeyByPrefix(ctx, prefix)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return mapAPIKeyToDomain(row), nil
}

func (r *APIKeyRepo) ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error) {
	q := r.store.getQueries(ctx)

	rows, err := q.ListAPIKeys(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.APIKey, len(rows))
	for i, row := range rows {
		result[i] = mapAPIKeyToDomain(row)
	}
	return result, nil
}

func (r *APIKeyRepo) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	q := r.store.getQueries(ctx)
	return q.UpdateAPIKeyLastUsed(ctx, id)
}

func (r *APIKeyRepo) DeleteAPIKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	q := r.store.getQueries(ctx)
	return q.DeleteAPIKey(ctx, sqlc.DeleteAPIKeyParams{
		ID:       id,
		TenantID: tenantID,
	})
}

func mapAPIKeyToDomain(row sqlc.ApiKeys) *domain.APIKey {
	return &domain.APIKey{
		ID:             row.ID,
		TenantID:       row.TenantID,
		Name:           row.Name,
		Type:           row.Type,
		Prefix:         row.Prefix,
		KeyHash:        row.KeyHash,
		Scopes:         row.Scopes,
		AllowedOrigins: row.AllowedOrigins,
		ExpiresAt:      row.ExpiresAt,
		CreatedAt:      row.CreatedAt,
	}
}