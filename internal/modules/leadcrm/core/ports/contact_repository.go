package ports

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"context"
)

type ContactRepository interface {
	Save(ctx context.Context, contact *domain.Contact) error
	FindByID(ctx context.Context, id string) (*domain.Contact, error)
}
