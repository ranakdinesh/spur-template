package ports

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"context"
)

type ActivityRepository interface {
	Save(ctx context.Context, activity *domain.Activity) error
}
