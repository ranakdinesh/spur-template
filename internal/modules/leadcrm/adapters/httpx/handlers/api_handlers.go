package handlers

import (
	"encoding/json"
	"net/http"
)

// IngestAPI handles POST /ingest/api/leads
// Headers:
// - Authorization: Bearer <API_KEY> (or X-API-Key)
// - Idempotency-Key: <UUID>
func (h *Handlers) IngestAPI(w http.ResponseWriter, r *http.Request) {
	// 1. Auth Stub - In real implementation, middleware or IdentityGateway call
	// For simplicity in this vertical slice, we'll assume a dummy tenant
	// token := r.Header.Get("Authorization") // Logic to validate
	tenantID := "stub-tenant-id"

	// 2. Idempotency Check Stub
	// idempotencyKey := r.Header.Get("Idempotency-Key")
	// if idempotencyKey != "" { ... check if processed ... }

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	leadID, err := h.Services.Lead.IngestFromAPI(r.Context(), tenantID, payload)
	if err != nil {
		// Log error
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"status":  "success",
		"lead_id": leadID,
	}
	json.NewEncoder(w).Encode(response)
}
