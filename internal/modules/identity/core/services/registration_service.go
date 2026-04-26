package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/modules/identity/core/workflows"
	"github.com/spurbase/spur/internal/platform/httpserver"
	"github.com/spurbase/spur/internal/platform/logger"
	"github.com/spurbase/spur/internal/platform/security"
	"github.com/spurbase/spur/internal/platform/temporal"
	"go.temporal.io/sdk/client"
)

// RegistrationService wires through port interfaces — no direct postgres dependency.
type RegistrationService struct {
	tenantRepo ports.TenantRepo
	userRepo   ports.UserRepo
	roleRepo   ports.RoleRepo
	moduleRepo ports.ModuleRepo
	permRepo   ports.PermissionRepo
	txManager  ports.TxManager
	temporal   *temporal.Client
	log        *logger.Loggerx
}

func NewRegistrationService(
	tenantRepo ports.TenantRepo,
	userRepo ports.UserRepo,
	roleRepo ports.RoleRepo,
	moduleRepo ports.ModuleRepo,
	permRepo ports.PermissionRepo,
	txManager ports.TxManager,
	temporalClient *temporal.Client,
	log *logger.Loggerx,
) *RegistrationService {
	return &RegistrationService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		moduleRepo: moduleRepo,
		permRepo:   permRepo,
		txManager:  txManager,
		temporal:   temporalClient,
		log:        log,
	}
}

// RegisterTenant creates Tenant + Admin User + TENANT_ADMIN role in one transaction.
func (s *RegistrationService) RegisterTenant(ctx context.Context, cmd ports.RegisterTenantCmd) (*ports.RegisteredTenantResult, error) {
	s.log.Info(ctx).Str("email", cmd.Email).Msg("Starting tenant registration")

	tenantName := cmd.CompanyName
	if tenantName == "" {
		tenantName = fmt.Sprintf("%s %s", cmd.FirstName, cmd.LastName)
	}

	passwordHash, err := security.HashPassword(cmd.Password)
	if err != nil {
		return nil, errors.New("internal server error")
	}

	var result ports.RegisteredTenantResult

	err = s.txManager.RunInTx(ctx, func(ctx context.Context) error {
		// 1. Tenant
		trialEnd := time.Now().Add(7 * 24 * time.Hour)
		subPlan := "trial"
		tenant := &domain.Tenant{
			ID:               uuid.New(),
			Name:             tenantName,
			Kind:             domain.TenantKindCustomer,
			TrialEndsAt:      &trialEnd,
			SubscriptionPlan: &subPlan,
		}
		t, err := s.tenantRepo.CreateTenant(ctx, tenant)
		if err != nil {
			return fmt.Errorf("create tenant: %w", err)
		}
		result.Tenant = t

		// 2. User
		userModel, err := domain.NewUser(t.ID, cmd.FirstName, cmd.LastName, cmd.Email, cmd.Mobile)
		if err != nil {
			return err
		}
		userModel.PasswordHash = passwordHash
		if cmd.AutoVerify {
			now := time.Now().UTC()
			userModel.VerifiedAt = &now
			userModel.EmailVerifiedAt = &now
			userModel.MobileVerifiedAt = &now
		} else {
			graceEnd := time.Now().Add(15 * 24 * time.Hour)
			userModel.VerificationGracePeriodEnd = &graceEnd
		}
		u, err := s.userRepo.CreateUser(ctx, userModel)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		result.User = u

		// 3. TENANT_ADMIN role
		roleCode := "TENANT_ADMIN"
		role, err := s.roleRepo.GetRoleByCode(ctx, roleCode, t.ID)
		if err != nil {
			role, err = s.roleRepo.CreateRole(ctx, &domain.Role{
				ID:          uuid.New(),
				TenantID:    t.ID,
				Name:        "Tenant Admin",
				Code:        &roleCode,
				Description: &roleCode,
				IsSystem:    true,
			})
			if err != nil {
				return fmt.Errorf("create admin role: %w", err)
			}
		}

		if err := s.roleRepo.AssignRoleToUser(ctx, u.ID, role.ID); err != nil {
			return fmt.Errorf("assign role: %w", err)
		}

		// 4. Enable modules and auto-grant permissions to TENANT_ADMIN
		for _, modID := range cmd.ModuleIDs {
			if err := s.moduleRepo.EnableModuleForTenant(ctx, t.ID, modID); err != nil {
				return fmt.Errorf("enable module %s: %w", modID, err)
			}
			perms, err := s.permRepo.ListPermissionsByModule(ctx, modID)
			if err == nil {
				for _, p := range perms {
					_ = s.permRepo.AssignPermissionToRole(ctx, role.ID, p.ID)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.startVerificationWorkflow(ctx, result.User.ID, result.Tenant.ID)
	return &result, nil
}

// CreateUser creates a user within an existing tenant.
func (s *RegistrationService) CreateUser(ctx context.Context, cmd ports.CreateUserCmd) (*domain.User, error) {
	if cmd.IsSuperAdmin {
		return nil, errors.New("cannot create additional Super Admin users")
	}
	for _, code := range cmd.Roles {
		if code == "SUPER_ADMIN" || code == "TENANT_ADMIN" {
			return nil, fmt.Errorf("role %s cannot be manually assigned", code)
		}
	}

	passwordHash, err := security.HashPassword(cmd.Password)
	if err != nil {
		return nil, err
	}

	var createdUser *domain.User

	err = s.txManager.RunInTx(ctx, func(ctx context.Context) error {
		u, err := domain.NewUser(cmd.TenantID, cmd.FirstName, cmd.LastName, cmd.Email, cmd.Mobile)
		if err != nil {
			return err
		}
		u.PasswordHash = passwordHash
		u.IsSuperAdmin = cmd.IsSuperAdmin
		graceEnd := time.Now().Add(15 * 24 * time.Hour)
		u.VerificationGracePeriodEnd = &graceEnd

		created, err := s.userRepo.CreateUser(ctx, u)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		createdUser = created

		roleCodes := cmd.Roles
		if len(roleCodes) == 0 {
			roleCodes = []string{"Member Login"}
		}
		for _, code := range roleCodes {
			role, err := s.roleRepo.GetRoleByCode(ctx, code, cmd.TenantID)
			if err != nil {
				role, err = s.roleRepo.GetSystemRoleByCode(ctx, code)
				if err != nil {
					return fmt.Errorf("role not found: %s", code)
				}
			}
			if err := s.roleRepo.AssignRoleToUser(ctx, created.ID, role.ID); err != nil {
				return fmt.Errorf("assign role %s: %w", code, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.startVerificationWorkflow(ctx, createdUser.ID, createdUser.TenantID)
	return createdUser, nil
}

func (s *RegistrationService) ListUsers(ctx context.Context, tenantID uuid.UUID) ([]*domain.User, error) {
	if httpserver.IsSuperAdmin(ctx) {
		return s.userRepo.ListAllUsers(ctx)
	}
	return s.userRepo.ListUsers(ctx, tenantID)
}

func (s *RegistrationService) UpdateUser(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, cmd ports.UpdateUserCmd) error {
	user, err := s.userRepo.GetUser(ctx, userID, tenantID)
	if err != nil {
		return err
	}
	if user.IsSuperAdmin {
		return errors.New("cannot modify Super Admin via this API")
	}

	return s.txManager.RunInTx(ctx, func(ctx context.Context) error {
		user.FirstName = cmd.FirstName
		user.LastName = cmd.LastName
		user.Mobile = cmd.Mobile

		if err := s.userRepo.UpdateUser(ctx, user); err != nil {
			return err
		}

		// Bump authz version so JWT caches invalidate
		user.AuthzVersion++
		if err := s.userRepo.UpdateUser(ctx, user); err != nil {
			return err
		}

		// Replace roles
		currentRoles, _ := s.roleRepo.ListUserRoles(ctx, userID)
		for _, r := range currentRoles {
			_ = s.roleRepo.RemoveRoleFromUser(ctx, userID, r.ID)
		}
		for _, code := range cmd.Roles {
			role, err := s.roleRepo.GetRoleByCode(ctx, code, tenantID)
			if err == nil {
				_ = s.roleRepo.AssignRoleToUser(ctx, userID, role.ID)
			}
		}
		return nil
	})
}

func (s *RegistrationService) UpdateUserLockStatus(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, isLocked bool) error {
	user, err := s.userRepo.GetUser(ctx, userID, tenantID)
	if err != nil {
		return err
	}
	if user.IsSuperAdmin {
		return errors.New("cannot lock Super Admin user")
	}
	return s.userRepo.UpdateUserLockStatus(ctx, userID, tenantID, isLocked)
}

func (s *RegistrationService) UpdateUserPassword(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, newPassword string) error {
	user, err := s.userRepo.GetUser(ctx, userID, tenantID)
	if err != nil {
		return err
	}
	if user.IsSuperAdmin {
		return errors.New("cannot change Super Admin password via this API")
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, userID, tenantID, hash)
}

func (s *RegistrationService) DeleteUser(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) error {
	user, err := s.userRepo.GetUser(ctx, userID, tenantID)
	if err != nil {
		return err
	}
	if user.IsSuperAdmin {
		return errors.New("cannot delete Super Admin user")
	}
	return s.userRepo.DeleteUser(ctx, userID, tenantID)
}

// startVerificationWorkflow fires the Temporal verification-reminder workflow
// if a Temporal client is configured. Errors are logged, not returned.
func (s *RegistrationService) startVerificationWorkflow(ctx context.Context, userID, tenantID uuid.UUID) {
	if s.temporal == nil {
		return
	}
	opts := client.StartWorkflowOptions{
		ID:        "verification_reminder_" + userID.String(),
		TaskQueue: "identity-queue",
	}
	_, err := s.temporal.ExecuteWorkflow(context.Background(), opts,
		workflows.VerificationReminderWorkflow,
		workflows.VerificationReminderInput{
			UserID:   userID.String(),
			TenantID: tenantID.String(),
		})
	if err != nil {
		s.log.Error(ctx).Err(err).Msg("failed to start verification workflow")
	}
}
