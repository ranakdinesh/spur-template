package sqlc

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type LeadForm struct {
	ID        string
	TenantID  string
	Name      string
	Slug      string
	Schema    json.RawMessage
	Settings  json.RawMessage
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LeadSubmission struct {
	ID          string
	TenantID    string
	FormID      string
	LeadID      sql.NullString
	SubmittedAt time.Time
	Payload     json.RawMessage
	IP          sql.NullString
	UserAgent   sql.NullString
}

type CreateLeadSubmissionParams struct {
	ID          string
	TenantID    string
	FormID      string
	LeadID      sql.NullString
	SubmittedAt time.Time
	Payload     json.RawMessage
	IP          sql.NullString
	UserAgent   sql.NullString
}

func (q *Queries) GetWebFormBySlug(ctx context.Context, slug string) (LeadForm, error) {
	return LeadForm{}, nil // Stub: Replace with real implementation if sqlc was running
}

func (q *Queries) CreateLeadSubmission(ctx context.Context, arg CreateLeadSubmissionParams) (string, error) {
	return "", nil // Stub
}
