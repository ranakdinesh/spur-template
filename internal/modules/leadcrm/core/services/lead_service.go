package services

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
	"time"
)

type LeadService struct {
	repo           ports.LeadRepository
	webFormRepo    ports.WebFormRepository
	activityRepo   ports.ActivityRepository
	engagementGate ports.EngagementGateway
}

func NewLeadService(repo ports.LeadRepository, webFormRepo ports.WebFormRepository, activityRepo ports.ActivityRepository, engagementGate ports.EngagementGateway) *LeadService {
	return &LeadService{
		repo:           repo,
		webFormRepo:    webFormRepo,
		activityRepo:   activityRepo,
		engagementGate: engagementGate,
	}
}

func (s *LeadService) CreateLead(ctx context.Context, lead *domain.Lead) error {
	return s.repo.Save(ctx, lead)
}

func (s *LeadService) CaptureFromWebform(ctx context.Context, slug string, payload map[string]interface{}, headers domain.SubmissionHeaders) error {
	// 1. Lookup form
	form, err := s.webFormRepo.FindBySlug(ctx, slug)
	if err != nil {
		return err
	}

	// 2. Spam check (Honeypot)
	if form.Settings.HoneypotField != "" {
		if val, ok := payload[form.Settings.HoneypotField]; ok && val != "" {
			// Honeypot filled, ignore silently or return error?
			// Return nil to fake success for bots
			return nil
		}
	}

	// 3. Normalize & Validate (Simplified)
	// In a real app we'd map fields based on schema.
	// For now taking basic fields assuming keys match
	// Also assume tenant_id from form
	lead := &domain.Lead{
		ID:        "gen-id-placeholder", // Should be generated
		TenantID:  form.TenantID,
		Source:    domain.SourceWebForm,
		Status:    domain.LeadStatusNew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 4. Save Lead
	if err := s.repo.Save(ctx, lead); err != nil {
		return err
	}

	// 5. Create Activity
	act := &domain.Activity{
		ID:        "act-id-placeholder",
		TenantID:  form.TenantID,
		Type:      domain.ActivityTypeNote, // Should be Capture
		Details:   "Lead captured via webform: " + form.Name,
		CreatedAt: time.Now(),
	}
	if err := s.activityRepo.Save(ctx, act); err != nil {
		// Log error but don't fail flow?
	}

	// 6. Save Submission
	sub := &domain.LeadSubmission{
		ID:          "sub-id-placeholder",
		TenantID:    form.TenantID,
		FormID:      form.ID,
		LeadID:      &lead.ID,
		SubmittedAt: time.Now(),
		Payload:     payload,
		Headers:     headers,
	}

	return s.webFormRepo.SaveSubmission(ctx, sub)
}

func (s *LeadService) IngestFromAPI(ctx context.Context, tenantID string, payload map[string]interface{}) (string, error) {
	// 1. Extract potential identities from payload
	email, _ := payload["email"].(string)
	phone, _ := payload["phone"].(string)

	var lead *domain.Lead
	var err error

	// 2. Try to find existing lead
	// Prefer Email, then Phone
	if email != "" {
		lead, err = s.repo.FindByIdentity(ctx, domain.IdentityTypeEmail, email)
		if err != nil && err != domain.ErrLeadNotFound {
			return "", err
		}
	}
	if lead == nil && phone != "" {
		lead, err = s.repo.FindByIdentity(ctx, domain.IdentityTypePhone, phone)
		if err != nil && err != domain.ErrLeadNotFound {
			return "", err
		}
	}

	// 3. Create or Update
	if lead == nil {
		// Create new
		lead = &domain.Lead{
			ID:        "gen-id-" + tenantID, // Stub
			TenantID:  tenantID,
			Source:    domain.SourceAPI,
			Status:    domain.LeadStatusNew,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	} else {
		// Update functionality stub
		lead.UpdatedAt = time.Now()
	}

	if err := s.repo.Save(ctx, lead); err != nil {
		return "", err
	}

	// 4. Link Identities
	if email != "" {
		// Ideally check if identity already exists to avoid dupes, implementation details left for finding
		_ = s.repo.AddIdentity(ctx, &domain.LeadIdentity{
			ID:       "ident-email-stub",
			TenantID: tenantID,
			LeadID:   lead.ID,
			Type:     domain.IdentityTypeEmail,
			Value:    email,
		})
	}

	// 5. Activity
	_ = s.activityRepo.Save(ctx, &domain.Activity{
		ID:        "act-api-ingest",
		TenantID:  tenantID,
		Type:      domain.ActivityTypeNote,
		Details:   "Lead ingested via API",
		CreatedAt: time.Now(),
	})

	return lead.ID, nil
}

func (s *LeadService) RequestEngagement(ctx context.Context, leadID string, instruction string) error {
	lead, err := s.repo.FindByID(ctx, leadID)
	if err != nil {
		return err
	}

	// Call Gateway
	jobID, err := s.engagementGate.EnqueueEngagement(ctx, lead, instruction)
	if err != nil {
		return err
	}

	// Record Activity
	_ = s.activityRepo.Save(ctx, &domain.Activity{
		ID:        "act-eng-" + lead.ID + "-" + jobID, // Stub
		TenantID:  lead.TenantID,
		Type:      domain.ActivityTypeEngagementRequested,
		Details:   "Engagement requested: " + instruction,
		CreatedAt: time.Now(),
		// Payload could store jobID
	})

	return nil
}
