package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/services"
	"github.com/spurbase/spur/internal/platform/httpserver"
)

type RBACHandler struct {
	svc *services.RBACService
}

func NewRBACHandler(svc *services.RBACService) *RBACHandler {
	return &RBACHandler{svc: svc}
}

// Tenants

func (h *RBACHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	t := &domain.Tenant{
		ID:   uuid.New(),
		Name: req.Name,
		Kind: domain.TenantKindCustomer,
	}

	result, err := h.svc.CreateTenant(r.Context(), t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *RBACHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.svc.ListTenants(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenants)
}

func (h *RBACHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	tenant, err := h.svc.GetTenant(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenant)
}

func (h *RBACHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	var req struct {
		Name             string `json:"name"`
		SubscriptionPlan string `json:"subscription_plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	t := &domain.Tenant{
		ID:               id,
		Name:             req.Name,
		SubscriptionPlan: &req.SubscriptionPlan,
	}

	if err := h.svc.UpdateTenant(r.Context(), t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RBACHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	if err := h.svc.DeleteTenant(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Modules

func (h *RBACHandler) CreateModule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	m := &domain.Module{
		ID:          uuid.New(),
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}

	result, err := h.svc.CreateModule(r.Context(), m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *RBACHandler) ListModules(w http.ResponseWriter, r *http.Request) {
	modules, err := h.svc.ListModules(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modules)
}

func (h *RBACHandler) ListTenantModules(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "id")
	tenantID, _ := uuid.Parse(tenantIDStr)

	modules, err := h.svc.ListTenantModules(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modules)
}

func (h *RBACHandler) EnableModuleForTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "id")
	tenantID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		ModuleID string `json:"module_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	modID, _ := uuid.Parse(req.ModuleID)

	if err := h.svc.EnableModuleForTenant(r.Context(), tenantID, modID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RBACHandler) DisableModuleForTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "id")
	tenantID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		ModuleID string `json:"module_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	modID, _ := uuid.Parse(req.ModuleID)

	if err := h.svc.DisableModuleForTenant(r.Context(), tenantID, modID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RBACHandler) GetModule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	module, err := h.svc.GetModule(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(module)
}

func (h *RBACHandler) UpdateModule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	m := &domain.Module{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.svc.UpdateModule(r.Context(), m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RBACHandler) DeleteModule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	if err := h.svc.DeleteModule(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Permissions

func (h *RBACHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ModuleID    string `json:"module_id"`
		Module      string `json:"module"`
		Key         string `json:"key"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	modID, _ := uuid.Parse(req.ModuleID)

	p := &domain.Permission{
		ID:          uuid.New(),
		ModuleID:    modID,
		Module:      req.Module,
		Key:         req.Key,
		Description: req.Description,
	}

	result, err := h.svc.CreatePermission(r.Context(), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *RBACHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var perms []*domain.Permission
	var err error

	if httpserver.IsSuperAdmin(ctx) {
		perms, err = h.svc.ListPermissions(ctx)
	} else {
		tenantIDStr := httpserver.GetTenantID(ctx)
		tenantID, _ := uuid.Parse(tenantIDStr)
		perms, err = h.svc.ListTenantAvailablePermissions(ctx, tenantID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(perms)
}

func (h *RBACHandler) ListPermissionsByModule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id") // Module ID
	id, _ := uuid.Parse(idStr)

	perms, err := h.svc.ListPermissionsByModule(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(perms)
}

func (h *RBACHandler) GetPermission(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	perm, err := h.svc.GetPermission(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(perm)
}

func (h *RBACHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	var req struct {
		Key         string `json:"key"`
		Description string `json:"description"`
		Module      string `json:"module"`
		ModuleID    string `json:"module_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	modID, _ := uuid.Parse(req.ModuleID)

	p := &domain.Permission{
		ID:          id,
		Key:         req.Key,
		Description: req.Description,
		Module:      req.Module,
		ModuleID:    modID,
	}

	if err := h.svc.UpdatePermission(r.Context(), p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RBACHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	if err := h.svc.DeletePermission(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Roles

func (h *RBACHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := httpserver.GetTenantID(r.Context())
	tenantID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		Name          string   `json:"name"`
		Code          *string  `json:"code"`
		Description   *string  `json:"description"`
		PermissionIDs []string `json:"permission_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	role := &domain.Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	// Create Role first
	createdRole, err := h.svc.CreateRole(r.Context(), role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Assign Permissions
	for _, permIDStr := range req.PermissionIDs {
		permID, err := uuid.Parse(permIDStr)
		if err == nil {
			// Best effort assignment, log error if fails but don't fail whole request?
			// Better to be transactional, but for now we do explicit calls.
			_ = h.svc.AssignPermissionToRole(r.Context(), createdRole.ID, permID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdRole)
}

func (h *RBACHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	isSA := httpserver.IsSuperAdmin(ctx)

	var roles []*domain.Role
	var err error

	if isSA {
		roles, err = h.svc.ListAllRoles(ctx)
	} else {
		tenantIDStr := httpserver.GetTenantID(ctx)
		tenantID, _ := uuid.Parse(tenantIDStr)
		roles, err = h.svc.ListRoles(ctx, tenantID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

func (h *RBACHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	var role *domain.Role
	var err error

	if httpserver.IsSuperAdmin(ctx) {
		// Use a helper or specific query that doesn't filter by tenant
		// For now, we can pass a special 'Nil' UUID or modify the service
		// Better: Add GetRoleByID to service
		role, err = h.svc.GetRoleByID(ctx, id)
	} else {
		tenantIDStr := httpserver.GetTenantID(ctx)
		tenantID, _ := uuid.Parse(tenantIDStr)
		role, err = h.svc.GetRole(ctx, id, tenantID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(role)
}

func (h *RBACHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	var req struct {
		Name        string `json:"name"`
		Code        *string `json:"code"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Fetch current role to get its tenant_id if Sa
	var tenantID uuid.UUID
	if httpserver.IsSuperAdmin(ctx) {
		existing, err := h.svc.GetRoleByID(ctx, id)
		if err != nil {
			http.Error(w, "role not found", http.StatusNotFound)
			return
		}
		tenantID = existing.TenantID
	} else {
		tenantIDStr := httpserver.GetTenantID(ctx)
		tenantID, _ = uuid.Parse(tenantIDStr)
	}

	role := &domain.Role{
		ID:          id,
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	if err := h.svc.UpdateRole(r.Context(), role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RBACHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	var tenantID uuid.UUID
	if httpserver.IsSuperAdmin(ctx) {
		existing, err := h.svc.GetRoleByID(ctx, id)
		if err != nil {
			http.Error(w, "role not found", http.StatusNotFound)
			return
		}
		tenantID = existing.TenantID
	} else {
		tenantIDStr := httpserver.GetTenantID(ctx)
		tenantID, _ = uuid.Parse(tenantIDStr)
	}

	// Protection: Super Admin role cannot be deleted
	role, err := h.svc.GetRole(ctx, id, tenantID)
	if err == nil && role.Name == "Super Admin" {
		http.Error(w, "cannot delete Super Admin role", http.StatusForbidden)
		return
	}

	if err := h.svc.DeleteRole(ctx, id, tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RBACHandler) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "roleID")
	roleID, _ := uuid.Parse(roleIDStr)

	// Protection: Super Admin role cannot be modified manually if it's the god role
	role, err := h.svc.GetRoleByID(r.Context(), roleID)
	if err == nil && role.Name == "Super Admin" {
		http.Error(w, "cannot modify Super Admin role permissions", http.StatusForbidden)
		return
	}

	var req struct {
		PermissionID string `json:"permission_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	permID, _ := uuid.Parse(req.PermissionID)

	if err := h.svc.AssignPermissionToRole(r.Context(), roleID, permID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RBACHandler) ListRolePermissions(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "roleID")
	roleID, _ := uuid.Parse(roleIDStr)

	perms, err := h.svc.ListRolePermissions(r.Context(), roleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(perms)
}

func (h *RBACHandler) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "roleID")
	roleID, _ := uuid.Parse(roleIDStr)

	// Protection: Super Admin role cannot be modified
	role, err := h.svc.GetRoleByID(r.Context(), roleID)
	if err == nil && role.Name == "Super Admin" {
		http.Error(w, "cannot modify Super Admin role permissions", http.StatusForbidden)
		return
	}

	permIDStr := chi.URLParam(r, "permissionID")
	permID, _ := uuid.Parse(permIDStr)

	if err := h.svc.RemovePermissionFromRole(r.Context(), roleID, permID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}