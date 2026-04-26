package domain

import (
	"time"
)

type Lead struct {
	ID                 string
	TenantID           string
	Source             Source
	Status             LeadStatus
	Stage              LeadStage
	Score              int
	QualifiedReason    string
	DisqualifiedReason string
	StageChangedAt     time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
