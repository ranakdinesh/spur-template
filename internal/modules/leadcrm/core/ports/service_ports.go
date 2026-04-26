package ports

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"context"
)

type LeadService interface {
	CreateLead(ctx context.Context, lead *domain.Lead) error
	CaptureFromWebform(ctx context.Context, slug string, payload map[string]interface{}, headers domain.SubmissionHeaders) error
	IngestFromAPI(ctx context.Context, tenantID string, payload map[string]interface{}) (string, error)
	RequestEngagement(ctx context.Context, leadID string, instruction string) error
}

type QualificationService interface {
	Qualify(ctx context.Context, leadID string) error
}

type ConversionService interface {
	Convert(ctx context.Context, leadID string, accountName string) error
}
