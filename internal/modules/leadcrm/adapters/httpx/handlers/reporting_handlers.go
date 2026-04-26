package handlers

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"github.com/spurbase/spur/internal/modules/leadcrm/views"
	"net/http"
	"time"

	"github.com/a-h/templ"
)

type ReportsHandler struct {
	repo ports.ReportingRepository
}

func NewReportsHandler(repo ports.ReportingRepository) *ReportsHandler {
	return &ReportsHandler{repo: repo}
}

func (h *ReportsHandler) ViewReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Hardcoded date range for v1
	end := time.Now()
	start := end.AddDate(0, -1, 0) // Last month

	funnel, err := h.repo.GetFunnelStats(ctx, start, end)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sources, err := h.repo.GetSourceStats(ctx)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	templ.Handler(views.Reports(funnel, sources)).ServeHTTP(w, r)
}
