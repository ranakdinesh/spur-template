package postgres

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
)

type AccountRepo struct {
	adapter *Adapter
}

func NewAccountRepo(adapter *Adapter) ports.AccountRepository {
	return &AccountRepo{adapter: adapter}
}

func (r *AccountRepo) Save(ctx context.Context, account *domain.Account) error {
	return nil
}

func (r *AccountRepo) FindByID(ctx context.Context, id string) (*domain.Account, error) {
	return nil, nil
}
