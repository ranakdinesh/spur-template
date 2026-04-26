package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/spurbase/spur/internal/modules/identity/core/services"
	"github.com/spurbase/spur/internal/platform/httpserver"
)

type APIKeyHandler struct {
	svc *services.APIKeyService
}

func NewAPIKeyHandler(svc *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{svc: svc}
}

func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := httpserver.GetTenantID(r.Context())
	tenantID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		Name           string   `json:"name"`
		Type           string   `json:"type"`
		Scopes         []string `json:"scopes"`
		AllowedOrigins []string `json:"allowed_origins"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = "secret" // Default
	}

	result, err := h.svc.CreateAPIKey(r.Context(), services.CreateAPIKeyInput{
		TenantID:       tenantID,
		Name:           req.Name,
		Type:           req.Type,
		Scopes:         req.Scopes,
		AllowedOrigins: req.AllowedOrigins,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := httpserver.GetTenantID(r.Context())
	tenantID, _ := uuid.Parse(tenantIDStr)

	keys, err := h.svc.ListAPIKeys(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := httpserver.GetTenantID(r.Context())
	tenantID, _ := uuid.Parse(tenantIDStr)

	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	if err := h.svc.DeleteAPIKey(r.Context(), id, tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
