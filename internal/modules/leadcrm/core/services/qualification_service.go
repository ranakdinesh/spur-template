package services

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
	"time"
)

type QualificationService struct {
	repo         ports.LeadRepository
	activityRepo ports.ActivityRepository
}

func NewQualificationService(repo ports.LeadRepository, activityRepo ports.ActivityRepository) *QualificationService {
	return &QualificationService{
		repo:         repo,
		activityRepo: activityRepo,
	}
}

func (s *QualificationService) Qualify(ctx context.Context, leadID string) error {
	lead, err := s.repo.FindByID(ctx, leadID)
	if err != nil {
		return err
	}

	// Attempt to enrich score based on what we know. (Stubbing email/company lookup)
	email := "" // TODO: Fetch from identities
	company := ""

	newScore := CalculateScore(lead, email, company)

	changes := false
	if lead.Score != newScore {
		// Emit Score Change
		_ = s.activityRepo.Save(ctx, &domain.Activity{
			ID:        "act-score-" + lead.ID, // Stub
			TenantID:  lead.TenantID,
			Type:      domain.ActivityTypeScoreChanged,
			Details:   "Score updated",
			CreatedAt: time.Now(),
		})
		lead.Score = newScore
		changes = true
	}

	newStage := EvaluateStage(lead.Stage, newScore)
	if lead.Stage != newStage {
		_ = s.activityRepo.Save(ctx, &domain.Activity{
			ID:        "act-stage-" + lead.ID, // Stub
			TenantID:  lead.TenantID,
			Type:      domain.ActivityTypeStageChanged,
			Details:   "Stage moved to " + string(newStage),
			CreatedAt: time.Now(),
		})
		lead.Stage = newStage
		lead.StageChangedAt = time.Now()
		changes = true
	}

	if changes {
		return s.repo.Save(ctx, lead)
	}
	return nil
}
