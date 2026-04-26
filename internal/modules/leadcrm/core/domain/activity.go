package domain

import (
	"time"
)

type Activity struct {
	ID        string
	TenantID  string
	Type      ActivityType
	Details   string
	CreatedAt time.Time
}
