package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
)

func (s *Store) CreateTenant(ctx context.Context, t *domain.Tenant) (*domain.Tenant, error) {
	var trialEndsAt pgtype.Timestamptz
	if t.TrialEndsAt != nil {
		trialEndsAt = pgtype.Timestamptz{Time: *t.TrialEndsAt, Valid: true}
	}

	arg := sqlc.CreateTenantParams{
		ID:               t.ID,
		Name:             t.Name,
		Kind:             sqlc.TenantKind(t.Kind),
		TrialEndsAt:      trialEndsAt,
		SubscriptionPlan: t.SubscriptionPlan,
	}

	row, err := s.getQueries(ctx).CreateTenant(ctx, arg)
	if err != nil {
		return nil, err
	}

	return mapTenantToDomain(row), nil
}

func (s *Store) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	row, err := s.getQueries(ctx).GetTenant(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapTenantToDomain(row), nil
}

func (s *Store) ListTenants(ctx context.Context) ([]*domain.Tenant, error) {
	rows, err := s.getQueries(ctx).ListTenants(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Tenant, len(rows))
	for i, r := range rows {
		result[i] = mapTenantToDomain(r)
	}
	return result, nil
}

func (s *Store) UpdateTenant(ctx context.Context, t *domain.Tenant) error {
	return s.getQueries(ctx).UpdateTenant(ctx, sqlc.UpdateTenantParams{
		ID:               t.ID,
		Name:             t.Name,
		SubscriptionPlan: t.SubscriptionPlan,
	})
}

func (s *Store) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return s.getQueries(ctx).DeleteTenant(ctx, id)
}

// Mapper
func mapTenantToDomain(row sqlc.Tenants) *domain.Tenant {
	t := &domain.Tenant{
		ID:        row.ID,
		Name:      row.Name,
		Kind:      domain.TenantKind(row.Kind),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.TrialEndsAt.Valid {
		val := row.TrialEndsAt.Time
		t.TrialEndsAt = &val
	}
	t.SubscriptionPlan = row.SubscriptionPlan
	return t
}
