package domain

import (
	"time"

	"github.com/google/uuid"
)

type ModuleStatus string

const (
	ModuleStatusActive   ModuleStatus = "active"
	ModuleStatusDisabled ModuleStatus = "disabled"
)

// Represents a global system module
type Module struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Represents a feature enabled for a specific tenant
type TenantModule struct {
	TenantID  uuid.UUID    `json:"tenant_id"`
	ModuleKey string       `json:"module_key"` // e.g. "crm"
	Status    ModuleStatus `json:"status"`
	EnabledAt time.Time    `json:"enabled_at"`
}
