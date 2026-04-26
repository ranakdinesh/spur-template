package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/platform/logger"
)

type RBACService struct {
	tenantRepo ports.TenantRepo
	userRepo   ports.UserRepo
	roleRepo   ports.RoleRepo
	moduleRepo ports.ModuleRepo
	permRepo   ports.PermissionRepo
	log        *logger.Loggerx
}

func NewRBACService(
	tenantRepo ports.TenantRepo,
	userRepo ports.UserRepo,
	roleRepo ports.RoleRepo,
	moduleRepo ports.ModuleRepo,
	permRepo ports.PermissionRepo,
	log *logger.Loggerx,
) *RBACService {
	return &RBACService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		moduleRepo: moduleRepo,
		permRepo:   permRepo,
		log:        log,
	}
}

// Tenants

func (s *RBACService) CreateTenant(ctx context.Context, t *domain.Tenant) (*domain.Tenant, error) {
	return s.tenantRepo.CreateTenant(ctx, t)
}

func (s *RBACService) ListTenants(ctx context.Context) ([]*domain.Tenant, error) {
	return s.tenantRepo.ListTenants(ctx)
}

func (s *RBACService) UpdateTenant(ctx context.Context, t *domain.Tenant) error {
	return s.tenantRepo.UpdateTenant(ctx, t)
}

func (s *RBACService) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return s.tenantRepo.DeleteTenant(ctx, id)
}

func (s *RBACService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return s.tenantRepo.GetTenant(ctx, id)
}

// Roles

func (s *RBACService) CreateRole(ctx context.Context, r *domain.Role) (*domain.Role, error) {
	return s.roleRepo.CreateRole(ctx, r)
}

func (s *RBACService) GetRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Role, error) {
	return s.roleRepo.GetRole(ctx, id, tenantID)
}

func (s *RBACService) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return s.roleRepo.GetRoleByID(ctx, id)
}

func (s *RBACService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	return s.roleRepo.ListRoles(ctx, tenantID)
}

func (s *RBACService) ListAllRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.roleRepo.ListAllRoles(ctx)
}

func (s *RBACService) UpdateRole(ctx context.Context, r *domain.Role) error {
	return s.roleRepo.UpdateRole(ctx, r)
}

func (s *RBACService) DeleteRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return s.roleRepo.DeleteRole(ctx, id, tenantID)
}

func (s *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.roleRepo.AssignRoleToUser(ctx, userID, roleID)
}

// Modules & Permissions

func (s *RBACService) CreateModule(ctx context.Context, m *domain.Module) (*domain.Module, error) {
	return s.moduleRepo.CreateModule(ctx, m)
}

func (s *RBACService) EnableModuleForTenant(ctx context.Context, tenantID, moduleID uuid.UUID) error {
	// 1. Enable module
	if err := s.moduleRepo.EnableModuleForTenant(ctx, tenantID, moduleID); err != nil {
		return err
	}

	// 2. Auto-assign all module permissions to TENANT_ADMIN
	role, err := s.roleRepo.GetRoleByCode(ctx, "TENANT_ADMIN", tenantID)
	if err != nil {
		s.log.Warn(ctx).Err(err).Str("tenant_id", tenantID.String()).Msg("TENANT_ADMIN role not found, skipping auto-permission assignment")
		return nil // Not fatal
	}

	perms, err := s.permRepo.ListPermissionsByModule(ctx, moduleID)
	if err != nil {
		return err
	}

	for _, p := range perms {
		_ = s.permRepo.AssignPermissionToRole(ctx, role.ID, p.ID)
	}

	return nil
}

func (s *RBACService) DisableModuleForTenant(ctx context.Context, tenantID, moduleID uuid.UUID) error {
	return s.moduleRepo.DisableModuleForTenant(ctx, tenantID, moduleID)
}

func (s *RBACService) GetModule(ctx context.Context, id uuid.UUID) (*domain.Module, error) {
	return s.moduleRepo.GetModule(ctx, id)
}

func (s *RBACService) ListModules(ctx context.Context) ([]*domain.Module, error) {
	return s.moduleRepo.ListModules(ctx)
}

func (s *RBACService) ListTenantModules(ctx context.Context, tenantID uuid.UUID) ([]*domain.Module, error) {
	return s.moduleRepo.ListTenantModules(ctx, tenantID)
}

func (s *RBACService) UpdateModule(ctx context.Context, m *domain.Module) error {
	return s.moduleRepo.UpdateModule(ctx, m)
}

func (s *RBACService) DeleteModule(ctx context.Context, id uuid.UUID) error {
	return s.moduleRepo.DeleteModule(ctx, id)
}

func (s *RBACService) CreatePermission(ctx context.Context, p *domain.Permission) (*domain.Permission, error) {
	return s.permRepo.CreatePermission(ctx, p)
}

func (s *RBACService) GetPermission(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	return s.permRepo.GetPermission(ctx, id)
}

func (s *RBACService) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return s.permRepo.ListPermissions(ctx)
}

func (s *RBACService) ListPermissionsByModule(ctx context.Context, moduleID uuid.UUID) ([]*domain.Permission, error) {
	return s.permRepo.ListPermissionsByModule(ctx, moduleID)
}

func (s *RBACService) ListTenantAvailablePermissions(ctx context.Context, tenantID uuid.UUID) ([]*domain.Permission, error) {
	// 1. Get Enabled Modules
	modules, err := s.moduleRepo.ListTenantModules(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch permissions for each module
	var allPerms []*domain.Permission
	for _, m := range modules {
		perms, err := s.permRepo.ListPermissionsByModule(ctx, m.ID)
		if err == nil {
			allPerms = append(allPerms, perms...)
		}
	}

	return allPerms, nil
}

func (s *RBACService) UpdatePermission(ctx context.Context, p *domain.Permission) error {
	return s.permRepo.UpdatePermission(ctx, p)
}

func (s *RBACService) DeletePermission(ctx context.Context, id uuid.UUID) error {
	return s.permRepo.DeletePermission(ctx, id)
}

func (s *RBACService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	// 1. Get Role
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	// 2. Get Permission
	perm, err := s.permRepo.GetPermission(ctx, permissionID)
	if err != nil {
		return err
	}

	// 3. Check if module is enabled for tenant
	enabledModules, err := s.moduleRepo.ListTenantModules(ctx, role.TenantID)
	if err != nil {
		return err
	}

	enabled := false
	for _, m := range enabledModules {
		if m.ID == perm.ModuleID {
			enabled = true
			break
		}
	}

	if !enabled {
		return fmt.Errorf("module '%s' is not enabled for this tenant", perm.Module)
	}

	return s.permRepo.AssignPermissionToRole(ctx, roleID, permissionID)
}

func (s *RBACService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.permRepo.RemovePermissionFromRole(ctx, roleID, permissionID)
}

func (s *RBACService) ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error) {
	return s.permRepo.ListRolePermissions(ctx, roleID)
}
