package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/platform/postgres"
)

// Modules

func (s *Store) CreateModule(ctx context.Context, m *domain.Module) (*domain.Module, error) {
	row, err := s.getQueries(ctx).CreateModule(ctx, sqlc.CreateModuleParams{
		ID:          m.ID,
		Code:        m.Code,
		Name:        m.Name,
		Description: postgres.ToPtr(m.Description),
	})
	if err != nil {
		return nil, err
	}
	return mapModuleToDomain(row), nil
}

func (s *Store) UpsertModule(ctx context.Context, m *domain.Module) (*domain.Module, error) {
	row, err := s.getQueries(ctx).UpsertModule(ctx, sqlc.UpsertModuleParams{
		ID:          m.ID,
		Code:        m.Code,
		Name:        m.Name,
		Description: postgres.ToPtr(m.Description),
	})
	if err != nil {
		return nil, err
	}
	return mapModuleToDomain(row), nil
}

func (s *Store) GetModule(ctx context.Context, id uuid.UUID) (*domain.Module, error) {
	row, err := s.getQueries(ctx).GetModule(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapModuleToDomain(row), nil
}

func (s *Store) ListModules(ctx context.Context) ([]*domain.Module, error) {
	rows, err := s.getQueries(ctx).ListModules(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Module, len(rows))
	for i, r := range rows {
		result[i] = mapModuleToDomain(r)
	}
	return result, nil
}

func (s *Store) UpdateModule(ctx context.Context, m *domain.Module) error {
	return s.getQueries(ctx).UpdateModule(ctx, sqlc.UpdateModuleParams{
		ID:          m.ID,
		Name:        m.Name,
		Description: postgres.ToPtr(m.Description),
	})
}

func (s *Store) DeleteModule(ctx context.Context, id uuid.UUID) error {
	return s.getQueries(ctx).DeleteModule(ctx, id)
}

// Tenant Modules

func (s *Store) EnableModuleForTenant(ctx context.Context, tenantID, moduleID uuid.UUID) error {
	return s.getQueries(ctx).EnableModuleForTenant(ctx, sqlc.EnableModuleForTenantParams{
		TenantID: tenantID,
		ModuleID: moduleID,
		Status:   string(domain.ModuleStatusActive),
	})
}

func (s *Store) DisableModuleForTenant(ctx context.Context, tenantID, moduleID uuid.UUID) error {
	return s.getQueries(ctx).DisableModuleForTenant(ctx, sqlc.DisableModuleForTenantParams{
		TenantID: tenantID,
		ModuleID: moduleID,
	})
}

func (s *Store) ListTenantModules(ctx context.Context, tenantID uuid.UUID) ([]*domain.Module, error) {
	rows, err := s.getQueries(ctx).ListTenantModules(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Module, len(rows))
	for i, r := range rows {
		result[i] = mapModuleToDomain(r)
	}
	return result, nil
}

// Permissions

func (s *Store) CreatePermission(ctx context.Context, p *domain.Permission) (*domain.Permission, error) {
	row, err := s.getQueries(ctx).CreatePermission(ctx, sqlc.CreatePermissionParams{
		ID:          p.ID,
		Key:         p.Key,
		Description: postgres.ToPtr(p.Description),
		Module:      p.Module,
		ModuleID:    pgtype.UUID{Bytes: p.ModuleID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapPermissionToDomain(row), nil
}

func (s *Store) UpsertPermission(ctx context.Context, p *domain.Permission) error {
	_, err := s.getQueries(ctx).UpsertPermission(ctx, sqlc.UpsertPermissionParams{
		ID:          p.ID,
		Key:         p.Key,
		Description: postgres.ToPtr(p.Description),
		Module:      p.Module,
		ModuleID:    pgtype.UUID{Bytes: p.ModuleID, Valid: true},
	})
	return err
}

func (s *Store) GetPermission(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	row, err := s.getQueries(ctx).GetPermission(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapPermissionToDomain(row), nil
}

func (s *Store) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	rows, err := s.getQueries(ctx).ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Permission, len(rows))
	for i, r := range rows {
		result[i] = mapPermissionToDomain(r)
	}
	return result, nil
}

func (s *Store) ListPermissionsByModule(ctx context.Context, moduleID uuid.UUID) ([]*domain.Permission, error) {
	rows, err := s.getQueries(ctx).ListPermissionsByModule(ctx, pgtype.UUID{Bytes: moduleID, Valid: true})
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Permission, len(rows))
	for i, r := range rows {
		result[i] = mapPermissionToDomain(r)
	}
	return result, nil
}

func (s *Store) UpdatePermission(ctx context.Context, p *domain.Permission) error {
	return s.getQueries(ctx).UpdatePermission(ctx, sqlc.UpdatePermissionParams{
		ID:          p.ID,
		Key:         p.Key,
		Description: postgres.ToPtr(p.Description),
		Module:      p.Module,
		ModuleID:    pgtype.UUID{Bytes: p.ModuleID, Valid: true},
	})
}

func (s *Store) DeletePermission(ctx context.Context, id uuid.UUID) error {
	return s.getQueries(ctx).DeletePermission(ctx, id)
}

func (s *Store) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.getQueries(ctx).AssignPermissionToRole(ctx, sqlc.AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (s *Store) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.getQueries(ctx).RemovePermissionFromRole(ctx, sqlc.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (s *Store) ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error) {
	rows, err := s.getQueries(ctx).ListRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Permission, len(rows))
	for i, r := range rows {
		result[i] = mapPermissionToDomain(r)
	}
	return result, nil
}

// Helpers

func mapModuleToDomain(row sqlc.Modules) *domain.Module {
	return &domain.Module{
		ID:          row.ID,
		Code:        row.Code,
		Name:        row.Name,
		Description: postgres.FromPtr(row.Description),
		CreatedAt:   row.CreatedAt,
	}
}

func mapPermissionToDomain(row sqlc.Permissions) *domain.Permission {
	return &domain.Permission{
		ID:          row.ID,
		ModuleID:    row.ModuleID.Bytes,
		Module:      row.Module,
		Key:         row.Key,
		Description: postgres.FromPtr(row.Description),
	}
}
