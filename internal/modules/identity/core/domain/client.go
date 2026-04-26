package domain

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID            string // Fosite expects string IDs (we can use UUID string)
	TenantID      uuid.UUID
	SecretHash    string   // For confidential clients (backend-to-backend)
	RedirectURIs  []string // Critical for Next.js callbacks
	GrantTypes    []string // e.g., "authorization_code", "refresh_token"
	ResponseTypes []string // e.g., "code"
	Scopes        []string // e.g., "openid", "profile", "offline"
	Audience      []string
	IsPublic      bool // True for Next.js (cannot hold secrets)
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Logic to check if this client can ask for a specific scope
func (c *Client) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}
