package ports

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"context"
)

type EnrichmentGateway interface {
	EnrichLead(ctx context.Context, lead *domain.Lead) error
}
