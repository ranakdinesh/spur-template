package domain

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"` // secret, publishable
	Prefix         string    `json:"prefix"`
	KeyHash        string    `json:"-"` // Store hash only!
	Scopes         []string  `json:"scopes"`
	IPAllowlist    []string  `json:"ip_allowlist"`
	AllowedOrigins []string  `json:"allowed_origins"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}
