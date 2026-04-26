package identity

import (
	"context"
	"crypto/rsa"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spurbase/spur/internal/modules/identity/adapters/communication"
	"github.com/spurbase/spur/internal/modules/identity/adapters/fosite_store"
	identityHttp "github.com/spurbase/spur/internal/modules/identity/adapters/http"
	"github.com/spurbase/spur/internal/modules/identity/adapters/http/handlers"
	"github.com/spurbase/spur/internal/modules/identity/adapters/oauth2"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/modules/identity/core/services"
	"github.com/spurbase/spur/internal/platform/httpserver"
	"github.com/spurbase/spur/internal/platform/logger"
	"github.com/spurbase/spur/internal/platform/temporal"
	"github.com/spurbase/spur/web"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds identity-specific settings populated from environment variables
// by internal/platform/config/config.go.
type Config struct {
	Issuer            string
	GlobalSecret      []byte
	JWTPrivateKeyPath string
	AuthClientId      uuid.UUID
	AuthClientSecret  string
}

// ─── Options ─────────────────────────────────────────────────────────────────

// Options is passed by app.go when constructing the identity module.
type Options struct {
	DB         *pgxpool.Pool
	Log        *logger.Loggerx
	Cfg        Config
	Renderer   *web.Renderer        // may be nil when using the Next.js UI
	AuthGuard  *httpserver.AuthGuard // pre-built by app.go
	PrivateKey *rsa.PrivateKey
	Temporal   *temporal.Client // nil = Temporal disabled
}

// ─── Services bundle ─────────────────────────────────────────────────────────

// Services exposes identity capabilities to app.go and other modules.
type Services struct {
	AuthService         ports.AuthService
	RegistrationService ports.RegistrationService
	FositeService       ports.FositeService
	RBACService         ports.RBACService
	APIKeyService       ports.APIKeyService
	ModuleService       *services.ModuleSyncService
}

// ─── Module ───────────────────────────────────────────────────────────────────

// Module is the public entry point for the identity module.
// app.go holds one reference and calls RegisterRoutes + Services.ModuleService.
type Module struct {
	Services  *Services
	Manifest  domain.Manifest
	PublicKey *rsa.PublicKey
}

// New wires the identity module from Options.
// Returns (nil, error) on failure — never calls log.Fatal or os.Exit.
func New(ctx context.Context, opt Options) (*Module, error) {
	if opt.DB == nil {
		return nil, fmt.Errorf("identity: DB pool is required")
	}
	if opt.PrivateKey == nil {
		return nil, fmt.Errorf("identity: PrivateKey is required")
	}
	if len(opt.Cfg.GlobalSecret) == 0 {
		return nil, fmt.Errorf("identity: GlobalSecret is required")
	}

	log := opt.Log

	// ── Postgres store (implements all repo + TxManager interfaces) ──────────
	store := postgres.NewStore(opt.DB)

	// ── Fosite OAuth2 provider ───────────────────────────────────────────────
	fositeStore := fosite_store.NewStore(store)
	provider, err := oauth2.NewProvider(
		fositeStore,
		log.Logger().With().Str("component", "fosite").Logger(),
		string(opt.Cfg.GlobalSecret),
		opt.Cfg.JWTPrivateKeyPath,
		opt.Cfg.Issuer,
	)
	if err != nil {
		return nil, fmt.Errorf("identity: oauth2 provider: %w", err)
	}

	// ── Communication (stub — wire Resend/Twilio via config) ────────────────
	commPort := communication.NewStubCommunicationService(log)

	// ── Repos ────────────────────────────────────────────────────────────────
	verificationRepo := postgres.NewVerificationRepo(store)
	apiKeyRepo := postgres.NewAPIKeyRepo(store)

	// ── Services ─────────────────────────────────────────────────────────────
	authSvc := services.NewAuthService(
		store, store, verificationRepo, store, store, commPort, log,
	)
	regSvc := services.NewRegistrationService(
		store, store, store, store, store, store, opt.Temporal, log,
	)
	fositeSvc := services.NewFositeService(store, provider)
	rbacSvc := services.NewRBACService(store, store, store, store, store, log)
	apiKeySvc := services.NewAPIKeyService(apiKeyRepo, log)
	moduleSvc := services.NewModuleSyncService(store, store, log)

	// ── Manifest: identity's own permissions ─────────────────────────────────
	manifest := domain.Manifest{
		Name:        "Identity & Access Management",
		Code:        "identity",
		Description: "Authentication, authorization, OAuth2, API keys, multi-tenancy.",
		Permissions: []domain.ManifestPermission{
			{Slug: "identity.admin.access", Description: "Access the identity admin panel"},
			{Slug: "identity.users.list", Description: "List users"},
			{Slug: "identity.users.create", Description: "Create users"},
			{Slug: "identity.users.view", Description: "View user details"},
			{Slug: "identity.users.update", Description: "Update users"},
			{Slug: "identity.users.delete", Description: "Delete users"},
			{Slug: "identity.roles.list", Description: "List roles"},
			{Slug: "identity.roles.create", Description: "Create roles"},
			{Slug: "identity.roles.update", Description: "Update roles"},
			{Slug: "identity.roles.delete", Description: "Delete roles"},
			{Slug: "identity.tenants.list", Description: "List tenants"},
			{Slug: "identity.tenants.create", Description: "Create tenants"},
			{Slug: "identity.tenants.update", Description: "Update tenants"},
			{Slug: "identity.tenants.delete", Description: "Delete tenants"},
			{Slug: "identity.apikeys.list", Description: "List API keys"},
			{Slug: "identity.apikeys.create", Description: "Create API keys"},
			{Slug: "identity.apikeys.delete", Description: "Revoke API keys"},
		},
	}

	return &Module{
		Services: &Services{
			AuthService:         authSvc,
			RegistrationService: regSvc,
			FositeService:       fositeSvc,
			RBACService:         rbacSvc,
			APIKeyService:       apiKeySvc,
			ModuleService:       moduleSvc,
		},
		Manifest:  manifest,
		PublicKey: &opt.PrivateKey.PublicKey,
	}, nil
}

// RegisterRoutes mounts all identity HTTP routes on the provided router.
func (m *Module) RegisterRoutes(r chi.Router) {
	authGuard, err := httpserver.NewAuthGuard(m.PublicKey)
	if err != nil {
		// Should never happen — PublicKey is validated in New()
		panic(fmt.Sprintf("identity: failed to build auth guard: %v", err))
	}

	verificationGuard := httpserver.NewVerificationGuard(
		&verificationProviderAdapter{authSvc: m.Services.AuthService},
	)

	clientHandler := handlers.NewClientHandler(m.Services.FositeService)
	regHandler := handlers.NewRegistrationHandler(m.Services.RegistrationService)
	authHandler := handlers.NewAuthHandler(m.Services.AuthService, m.Services.FositeService)
	apiKeyHandler := handlers.NewAPIKeyHandler(m.Services.APIKeyService)
	rbacHandler := handlers.NewRBACHandler(m.Services.RBACService)

	identityHttp.RegisterRoutes(
		r,
		clientHandler,
		regHandler,
		authHandler,
		verificationGuard,
		authGuard,
		apiKeyHandler,
		rbacHandler,
	)
}

// ─── Internal adapters ────────────────────────────────────────────────────────

// verificationProviderAdapter bridges ports.AuthService → httpserver.VerificationProvider.
type verificationProviderAdapter struct {
	authSvc ports.AuthService
}

func (a *verificationProviderAdapter) GetVerificationStatus(
	ctx context.Context,
	userID uuid.UUID,
	tenantID uuid.UUID,
) (*httpserver.VerificationStatus, error) {
	status, err := a.authSvc.GetVerificationStatus(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	return &httpserver.VerificationStatus{
		IsVerified:         status.IsVerified,
		GracePeriodExpired: status.GracePeriodExpired,
	}, nil
}
