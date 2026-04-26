package ports

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"context"
)

type WebFormRepository interface {
	FindBySlug(ctx context.Context, slug string) (*domain.LeadForm, error)
	SaveSubmission(ctx context.Context, submission *domain.LeadSubmission) error
}
