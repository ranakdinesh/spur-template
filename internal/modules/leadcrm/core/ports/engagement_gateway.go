package ports

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"context"
)

type EngagementGateway interface {
	// EnqueueEngagement enqueues a job (e.g. via Temporal) to engage the lead.
	// Returns the external JobID/WorkflowID.
	EnqueueEngagement(ctx context.Context, lead *domain.Lead, instruction string) (string, error)
}
