package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) EngageLead(w http.ResponseWriter, r *http.Request) {
	leadID := chi.URLParam(r, "id")
	if leadID == "" {
		http.Error(w, "lead id required", http.StatusBadRequest)
		return
	}

	var payload struct {
		Instruction string `json:"instruction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if payload.Instruction == "" {
		http.Error(w, "instruction required", http.StatusBadRequest)
		return
	}

	if err := h.Services.Lead.RequestEngagement(r.Context(), leadID, payload.Instruction); err != nil {
		// Log err
		http.Error(w, "failed to request engagement", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"queued"}`))
}
