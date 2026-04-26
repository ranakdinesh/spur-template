package domain

import "github.com/google/uuid"

type Permission struct {
	ID          uuid.UUID `json:"id"`
	ModuleID    uuid.UUID `json:"module_id"`
	Module      string    `json:"module"` // e.g., "identity"
	Key         string    `json:"key"`    // e.g., "tenants.create"
	Description string    `json:"description"`
}

// FullKey returns the complete permission string, e.g., "identity.tenants.create"
func (p Permission) FullKey() string {
	return p.Module + "." + p.Key
}
