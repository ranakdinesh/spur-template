package postgres

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type WebFormRepo struct {
	adapter *Adapter
}

func NewWebFormRepo(adapter *Adapter) ports.WebFormRepository {
	return &WebFormRepo{adapter: adapter}
}

func (r *WebFormRepo) FindBySlug(ctx context.Context, slug string) (*domain.LeadForm, error) {
	// In a real scenario, we'd use r.adapter.q.GetWebFormBySlug matching the tenant context derived or passed
	// For now, returning a mock or stub based on the generated code
	// Assuming generated code signature matches somewhat.

	// Stub implementation using SQLC Generated struct
	row, err := r.adapter.q.GetWebFormBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrWebFormNotFound
		}
		return nil, err
	}

	var schema domain.FormSchema
	if err := json.Unmarshal(row.Schema, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	var settings domain.FormSettings
	if err := json.Unmarshal(row.Settings, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &domain.LeadForm{
		ID:        row.ID,
		TenantID:  row.TenantID,
		Name:      row.Name,
		Slug:      row.Slug,
		Schema:    schema,
		Settings:  settings,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *WebFormRepo) SaveSubmission(ctx context.Context, sub *domain.LeadSubmission) error {
	payloadBytes, err := json.Marshal(sub.Payload)
	if err != nil {
		return err
	}

	var leadID sql.NullString
	if sub.LeadID != nil {
		leadID = sql.NullString{String: *sub.LeadID, Valid: true}
	}

	// Assuming stubs match
	_, err = r.adapter.q.CreateLeadSubmission(ctx, sqlc.CreateLeadSubmissionParams{
		ID:          sub.ID,
		TenantID:    sub.TenantID,
		FormID:      sub.FormID,
		LeadID:      leadID,
		SubmittedAt: sub.SubmittedAt,
		Payload:     payloadBytes,
		IP:          sql.NullString{String: sub.Headers.IP, Valid: sub.Headers.IP != ""},
		UserAgent:   sql.NullString{String: sub.Headers.UserAgent, Valid: sub.Headers.UserAgent != ""},
	})
	return err
}
