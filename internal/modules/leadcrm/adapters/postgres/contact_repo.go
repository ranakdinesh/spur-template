package postgres

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
)

type ContactRepo struct {
	adapter *Adapter
}

func NewContactRepo(adapter *Adapter) ports.ContactRepository {
	return &ContactRepo{adapter: adapter}
}

func (r *ContactRepo) Save(ctx context.Context, contact *domain.Contact) error {
	return nil
}

func (r *ContactRepo) FindByID(ctx context.Context, id string) (*domain.Contact, error) {
	return nil, nil
}
