package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/platform/appmodule"
	"github.com/spurbase/spur/internal/platform/logger"
)

type ModuleSyncService struct {
	moduleRepo ports.ModuleRepo
	permRepo   ports.PermissionRepo
	log        *logger.Loggerx
}

func NewModuleSyncService(
	moduleRepo ports.ModuleRepo,
	permRepo ports.PermissionRepo,
	log *logger.Loggerx,
) *ModuleSyncService {
	return &ModuleSyncService{
		moduleRepo: moduleRepo,
		permRepo:   permRepo,
		log:        log,
	}
}

func (s *ModuleSyncService) SyncModules(ctx context.Context, modules []appmodule.Module) error {
	s.log.Info(ctx).Msg("Starting module and permission synchronization")

	for _, m := range modules {
		def := m.GetDefinition()
		s.log.Info(ctx).Str("module", def.Code).Msg("Syncing module")

		// 1. Create or Update Module
		// We need an Upsert capability. If CreateModule fails with conflict, we should Update.
		// For now, let's try Create, if error, Update (or rely on Repo Upsert if implemented).
		// The current Repo CreateModule is simple INSERT.
		// To fix this PROPERLY without changing Repo interface too much, we can implement Upsert in the Repo layer.
		// OR check existence.
		
		// Ideally, we add UpsertModule to the ModuleRepo interface.
		// But let's use the existing methods logic if possible or extend.
		
		// IMPORTANT: The previous sync.go used `s.store.Queries.UpsertModule`.
		// Since we are now in the Service layer using Ports (Interfaces), we cannot access `s.store.Queries` directly.
		// We must extend the interface `ModuleRepo` and `PermissionRepo` to support Upsert, 
		// OR handle the logic here (Get -> if nil Create else Update).
		
		// Let's go with the Check-Then-Act approach for now to avoid interface breakage if not needed,
		// though Upsert is atomic and better.
		
		// Wait, the previous `rbac.sql` added UpsertModule. 
		// So `postgres.Store` has `UpsertModule`.
		// We should add `UpsertModule` to `ports.ModuleRepo`.
		
		mod := &domain.Module{
			ID:          uuid.New(), // ID is ignored on update, used on insert
			Code:        def.Code,
			Name:        def.Name,
			Description: def.Description,
		}
		
		// Calling generic Upsert method (we need to add this to interface)
		// For now, let's assume we will add `UpsertModule` to ports.
		savedMod, err := s.moduleRepo.UpsertModule(ctx, mod)
		if err != nil {
			return fmt.Errorf("failed to upsert module %s: %w", def.Code, err)
		}

		// 2. Sync Permissions
		for _, p := range def.Permissions {
			perm := &domain.Permission{
				ID:          uuid.New(),
				Key:         p.Key,
				Description: p.Description,
				Module:      def.Code,
				ModuleID:    savedMod.ID,
			}
			
			if err := s.permRepo.UpsertPermission(ctx, perm); err != nil {
				return fmt.Errorf("failed to upsert permission %s.%s: %w", def.Code, p.Key, err)
			}
		}
	}

	s.log.Info(ctx).Msg("Synchronization completed successfully")
	return nil
}

// RegisterManifest upserts a module record and all its declared permissions.
// Called by app.go once per module at startup.
func (s *ModuleSyncService) RegisterManifest(ctx context.Context, m domain.Manifest) error {
	s.log.Info(ctx).Str("module", m.Code).Msg("Registering module manifest")

	mod := &domain.Module{
		ID:          uuid.New(),
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
	}

	savedMod, err := s.moduleRepo.UpsertModule(ctx, mod)
	if err != nil {
		return fmt.Errorf("upsert module %s: %w", m.Code, err)
	}

	for _, p := range m.Permissions {
		// Slug format: "module.resource.action" — Key stored without module prefix
		key := p.Slug
		perm := &domain.Permission{
			ID:          uuid.New(),
			Key:         key,
			Module:      m.Code,
			ModuleID:    savedMod.ID,
			Description: p.Description,
		}
		if err := s.permRepo.UpsertPermission(ctx, perm); err != nil {
			return fmt.Errorf("upsert permission %s: %w", p.Slug, err)
		}
	}

	s.log.Info(ctx).Str("module", m.Code).Int("permissions", len(m.Permissions)).Msg("Manifest registered")
	return nil
}
