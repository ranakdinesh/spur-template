package postgres

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
)

type LeadRepo struct {
	adapter *Adapter
}

func NewLeadRepo(adapter *Adapter) ports.LeadRepository {
	return &LeadRepo{adapter: adapter}
}

func (r *LeadRepo) Save(ctx context.Context, lead *domain.Lead) error {
	return nil
}

func (r *LeadRepo) FindByID(ctx context.Context, id string) (*domain.Lead, error) {
	return nil, nil
}

func (r *LeadRepo) FindByIdentity(ctx context.Context, idType domain.LeadIdentityType, value string) (*domain.Lead, error) {
	// Stub implementation
	// row, err := r.adapter.q.FindLeadByIdentity(ctx, sqlc.FindLeadByIdentityParams{...})
	// For now return dummy or error
	return nil, domain.ErrLeadNotFound
}

func (r *LeadRepo) AddIdentity(ctx context.Context, identity *domain.LeadIdentity) error {
	// Stub implementation
	// r.adapter.q.CreateLeadIdentity(ctx, identity.ToParams())
	return nil
}

func (r *LeadRepo) GetForUpdate(ctx context.Context, id string) (*domain.Lead, error) {
	// Stub implementation: In real life this would run queries.sql GetLeadForUpdate
	// SELECT * FROM leads WHERE id = $1 FOR UPDATE
	return r.FindByID(ctx, id)
}
