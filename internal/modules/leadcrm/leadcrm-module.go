package leadcrm

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/adapters/httpx"
	"github.com/spurbase/spur/internal/modules/leadcrm/adapters/httpx/handlers"
	"github.com/spurbase/spur/internal/modules/leadcrm/adapters/postgres"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/services"
	"github.com/spurbase/spur/internal/platform/logger"
	"github.com/spurbase/spur/web"
)

type Options struct {
	DB       *pgxpool.Pool
	Log      *logger.Loggerx
	Renderer *web.Renderer
}

type Module struct {
	Services *services.Services
	Handlers *handlers.Handlers
	Manifest domain.Manifest
}

func New(ctx context.Context, opt Options) (*Module, error) {
	pgAdapter := postgres.New(opt.DB)

	leadRepo := postgres.NewLeadRepo(pgAdapter)
	contactRepo := postgres.NewContactRepo(pgAdapter)
	accountRepo := postgres.NewAccountRepo(pgAdapter)
	activityRepo := postgres.NewActivityRepo(pgAdapter)
	webFormRepo := postgres.NewWebFormRepo(pgAdapter)
	engagementGate := postgres.NewEngagementGateway()

	leadService := services.NewLeadService(leadRepo, webFormRepo, activityRepo, engagementGate)
	qualService := services.NewQualificationService(leadRepo, activityRepo)
	convService := services.NewConversionService(leadRepo, contactRepo, accountRepo, activityRepo, pgAdapter) // pgAdapter implements TransactionManager

	reportingRepo := postgres.NewReportingRepo(pgAdapter)
	reportsHandler := handlers.NewReportsHandler(reportingRepo)

	svc := services.New(leadService, qualService, convService)
	h := handlers.New(svc, reportsHandler)

	manifest := domain.Manifest{
		Name:        "Lead CRM",
		Code:        "leadcrm",
		Description: "Manage leads, webforms, and CRM reports.",
		Permissions: []domain.ManifestPermission{
			{Slug: "leadcrm.leads.view", Description: ptr("View leads")},
			{Slug: "leadcrm.leads.create", Description: ptr("Create leads")},
			{Slug: "leadcrm.leads.update", Description: ptr("Update leads")},
			{Slug: "leadcrm.webforms.view", Description: ptr("View webforms")},
			{Slug: "leadcrm.webforms.create", Description: ptr("Create webforms")},
			{Slug: "leadcrm.reports.view", Description: ptr("View CRM reports")},
		},
	}

	return &Module{
		Services: svc,
		Handlers: h,
		Manifest: manifest,
	}, nil
}

func (m *Module) RegisterRoutes(r chi.Router) {
	httpx.Mount(r, m.Handlers)
}

func ptr(s string) *string {
	return &s
}
