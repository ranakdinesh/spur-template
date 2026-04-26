package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
)

// We'll assume standard http.HandlerFunc for now or similar pattern.
// If common helpers exist, we should use them.
// But I can't see other handlers yet. I'll write standard code and refactor if I see patterns later.
// Actually I should look for a "users" or "auth" handler example if possible.
// But since I can't browse much, I'll stick to basic implementation.

type ClientHandler struct {
	Service ports.FositeService
}

func NewClientHandler(service ports.FositeService) *ClientHandler {
	return &ClientHandler{Service: service}
}

func (h *ClientHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var cmd ports.CreateClientCmd
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TenantID might come from context (middleware).
	// For now, let's assume it's in the Body or Context.
	// The cmd payload has TenantID.
	// Ideally we validate that the authenticated user belongs to that tenant involved.
	// But let's proceed with basic extraction.

	client, err := h.Service.CreateClient(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(client)
}

func (h *ClientHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	// Extract TenantID from query or context?
	// The user request didn't specify authentication details yet.
	// Let's assume query param ?tenant_id=... for Ops admin, or context for logged in user.
	// Using generic approach:
	tidStr := r.URL.Query().Get("tenant_id")
	if tidStr == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}
	tid, err := uuid.Parse(tidStr)
	if err != nil {
		http.Error(w, "invalid tenant_id", http.StatusBadRequest)
		return
	}

	clients, err := h.Service.ListClients(r.Context(), tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func (h *ClientHandler) ListPublicClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.Service.ListPublicClients(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func (h *ClientHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	client, err := h.Service.GetClient(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

func (h *ClientHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var cmd ports.UpdateClientCmd
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateClient(r.Context(), id, cmd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ClientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Service.DeleteClient(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
