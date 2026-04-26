package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/platform/logger"
)

type RoleService struct {
	store *postgres.Store
	log   *logger.Loggerx
}

func NewRoleService(store *postgres.Store, log *logger.Loggerx) *RoleService {
	return &RoleService{store: store, log: log}
}

func (s *RoleService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	return s.store.ListRoles(ctx, tenantID)
}

func (s *RoleService) CreateRole(ctx context.Context, tenantID uuid.UUID, name, code string) (*domain.Role, error) {
	roleModel := &domain.Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Code:        &code,
		Description: &name,
	}
	return s.store.CreateRole(ctx, roleModel)
}
