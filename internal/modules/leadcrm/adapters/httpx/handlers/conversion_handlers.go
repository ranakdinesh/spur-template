package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) ConvertLead(w http.ResponseWriter, r *http.Request) {
	leadID := chi.URLParam(r, "id")
	if leadID == "" {
		http.Error(w, "lead id required", http.StatusBadRequest)
		return
	}

	var payload struct {
		AccountName string `json:"account_name"`
	}
	// Ignore error if body is empty, just take default
	_ = json.NewDecoder(r.Body).Decode(&payload)

	if err := h.Services.Conversion.Convert(r.Context(), leadID, payload.AccountName); err != nil {
		// Log err
		http.Error(w, "failed to convert lead", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"converted"}`))
}
