package httpx

import (
	"github.com/go-chi/chi/v5"
	"github.com/spurbase/spur/internal/modules/leadcrm/adapters/httpx/handlers"
	"github.com/spurbase/spur/internal/platform/httpserver"
)

func Mount(r chi.Router, h *handlers.Handlers) {
	// Public routes
	r.Post("/ingest/webform/{slug}", h.IngestWebform)
	r.Post("/ingest/api/leads", h.IngestAPI)

	// Protected routes
	r.Route("/app/crm", func(r chi.Router) {
		// We assume AuthGuard and RBAC middleware are applied at the app level or parent router
		// But if we need specific permissions, we can add them here.
		
		r.Route("/leads", func(r chi.Router) {
			r.Use(httpserver.RequirePermission("leadcrm.leads.view"))
			// r.Get("/", h.ListLeads)
			
			r.Post("/{id}/engage", h.EngageLead)
			r.Post("/{id}/convert", h.ConvertLead)
		})

		r.Get("/reports", h.Reports.ViewReports)
		
		// Webforms management
		r.Route("/webforms", func(r chi.Router) {
			r.Use(httpserver.RequirePermission("leadcrm.webforms.view"))
			// r.Get("/", h.ListWebforms)
		})
	})
}
