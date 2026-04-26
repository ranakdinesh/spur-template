package postgres

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
)

type ActivityRepo struct {
	adapter *Adapter
}

func NewActivityRepo(adapter *Adapter) ports.ActivityRepository {
	return &ActivityRepo{adapter: adapter}
}

func (r *ActivityRepo) Save(ctx context.Context, activity *domain.Activity) error {
	return nil
}
