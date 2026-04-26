package handlers

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) IngestWebform(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "slug required", http.StatusBadRequest)
		return
	}

	// Parse payload
	var payload map[string]interface{}

	// Support both JSON and Form encodings
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		payload = make(map[string]interface{})
		for k, v := range r.PostForm {
			if len(v) > 0 {
				payload[k] = v[0]
			}
		}
	}

	headers := domain.SubmissionHeaders{
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}

	if err := h.Services.Lead.CaptureFromWebform(r.Context(), slug, payload, headers); err != nil {
		if err == domain.ErrWebFormNotFound {
			http.Error(w, "form not found", http.StatusNotFound)
			return
		}
		// Log error
		http.Error(w, "internal processing error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"captured"}`))
}
