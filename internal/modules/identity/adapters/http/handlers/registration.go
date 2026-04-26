package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/platform/httpserver"
)

type RegistrationHandler struct {
	Service ports.RegistrationService
}

func NewRegistrationHandler(service ports.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{Service: service}
}

// RegisterTenant - Public Self-Registration
func (h *RegistrationHandler) RegisterTenant(w http.ResponseWriter, r *http.Request) {
	var cmd ports.RegisterTenantCmd
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if cmd.Email == "" || cmd.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.Service.RegisterTenant(r.Context(), cmd)
	if err != nil {
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// The result contains *domain.Tenant and *domain.User, which should serialize to JSON correctly
	json.NewEncoder(w).Encode(result)
}

// RegisterUser - Admin creates user (Ops or Tenant Admin)
func (h *RegistrationHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var cmd ports.CreateUserCmd
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("RegisterUser: Received CMD: %+v\n", cmd)

	// Injected: Ensure tenant_id is set from the context if not provided
	if cmd.TenantID == uuid.Nil {
		tenantIDStr := httpserver.GetTenantID(r.Context())
		if tenantIDStr != "" {
			cmd.TenantID, _ = uuid.Parse(tenantIDStr)
		}
	}

	user, err := h.Service.CreateUser(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *RegistrationHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr := httpserver.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)

	users, err := h.Service.ListUsers(ctx, tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *RegistrationHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := chi.URLParam(r, "id")
	userID, _ := uuid.Parse(userIDStr)

	tenantIDStr := httpserver.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)

	var cmd ports.UpdateUserCmd
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateUser(ctx, userID, tenantID, cmd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RegistrationHandler) UpdateUserLockStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := chi.URLParam(r, "id")
	userID, _ := uuid.Parse(userIDStr)

	tenantIDStr := httpserver.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		IsLocked bool `json:"is_locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateUserLockStatus(ctx, userID, tenantID, req.IsLocked); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RegistrationHandler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := chi.URLParam(r, "id")
	userID, _ := uuid.Parse(userIDStr)

	tenantIDStr := httpserver.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateUserPassword(ctx, userID, tenantID, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RegistrationHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := chi.URLParam(r, "id")
	userID, _ := uuid.Parse(userIDStr)

	tenantIDStr := httpserver.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if err := h.Service.DeleteUser(ctx, userID, tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
