package services

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"strings"
)

// CalculateScore computes the lead score based on its data.
// This is a pure function.
func CalculateScore(lead *domain.Lead, email string, company string) int {
	score := 0

	// Source Rules
	if lead.Source == domain.SourceWebForm {
		score += 5
	}

	// Data Completion Rules
	if email != "" {
		score += 10
		if !strings.HasSuffix(email, "@gmail.com") && !strings.HasSuffix(email, "@yahoo.com") && !strings.HasSuffix(email, "@hotmail.com") {
			// B2B Email Guess
			score += 50
		}
	}

	if company != "" {
		score += 20
	}

	// Engagement Rules (Stub - could loop through activities if available)

	return score
}

// EvaluateStage determines the next stage based on score and status.
func EvaluateStage(currentStage domain.LeadStage, currentScore int) domain.LeadStage {
	// Simple threshold rules
	if currentScore >= 80 {
		return domain.StageQualify
	}
	if currentScore >= 30 {
		return domain.StageTriage
	}
	return domain.StageCapture
}
