package ports

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"context"
)

type AccountRepository interface {
	Save(ctx context.Context, account *domain.Account) error
	FindByID(ctx context.Context, id string) (*domain.Account, error)
}
