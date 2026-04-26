package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TenantKind string

const (
	TenantKindOps      TenantKind = "ops"
	TenantKindCustomer TenantKind = "customer"
)

var (
	ErrInvalidTenantName = errors.New("invalid tenant name")
)

type Tenant struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	Kind             TenantKind `json:"kind"`
	TrialEndsAt      *time.Time `json:"trial_ends_at,omitempty"`
	SubscriptionPlan *string    `json:"subscription_plan,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func NewTenant(name string, kind TenantKind) (*Tenant, error) {
	if name == "" {
		return nil, ErrInvalidTenantName
	}

	return &Tenant{
		ID:        uuid.New(),
		Name:      name,
		Kind:      kind,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}
