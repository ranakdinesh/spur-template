package domain

import (
	"time"
)

type Contact struct {
	ID        string
	TenantID  string
	FirstName string
	LastName  string
	Email     string
	LeadID    *string // Optional link to origin lead
	CreatedAt time.Time
	UpdatedAt time.Time
}
