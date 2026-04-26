package ports

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"context"
)

type LeadRepository interface {
	Save(ctx context.Context, lead *domain.Lead) error
	FindByID(ctx context.Context, id string) (*domain.Lead, error)
	FindByIdentity(ctx context.Context, idType domain.LeadIdentityType, value string) (*domain.Lead, error)
	AddIdentity(ctx context.Context, identity *domain.LeadIdentity) error
	GetForUpdate(ctx context.Context, id string) (*domain.Lead, error)
}
