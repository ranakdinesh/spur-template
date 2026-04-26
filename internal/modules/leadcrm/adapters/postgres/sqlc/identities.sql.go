package sqlc

import (
	"context"
)

type CreateLeadIdentityParams struct {
	ID       string
	TenantID string
	LeadID   string
	Type     string
	Value    string
}

type FindLeadByIdentityParams struct {
	TenantID string
	Type     string
	Value    string
}

func (q *Queries) CreateLeadIdentity(ctx context.Context, arg CreateLeadIdentityParams) error {
	return nil // Stub
}

func (q *Queries) FindLeadByIdentity(ctx context.Context, arg FindLeadByIdentityParams) (Lead, error) {
	return Lead{}, nil // Stub
}
