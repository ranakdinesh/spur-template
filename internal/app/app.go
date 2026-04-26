package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity"
	"github.com/spurbase/spur/internal/modules/leadcrm"
	"github.com/spurbase/spur/internal/platform/config"
	"github.com/spurbase/spur/internal/platform/db"
	"github.com/spurbase/spur/internal/platform/httpserver"
	"github.com/spurbase/spur/internal/platform/logger"
	"github.com/spurbase/spur/internal/ui/pages"
	"github.com/spurbase/spur/web"
	// Ensure this matches your folder name (logger vs logging)
)

type App struct {
	Config   config.Config
	Logger   *logger.Loggerx
	Identity *identity.Module
	LeadCRM  *leadcrm.Module
	Server   *httpserver.Server
}

// New initializes the system. It fails fast (returns error) if something critical is missing.
func New(ctx context.Context) (*App, error) {
	// 1. Load Configuration
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		return nil, fmt.Errorf("app: failed to load config: %w", err)
	}

	// 2. Initialize Logging
	log := logger.NewWithOptions(logger.Options{
		Environment: cfg.AppEnv,
	})
	// setting up the database
	dbPool := db.NewPool(ctx, cfg.DatabaseURL)
	renderer := web.NewRenderer()

	privateKey, err := config.LoadPrivateKey(cfg.JWTPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("app: failed to load private key: %w", err)
	}

	authclientid := uuid.MustParse(cfg.AuthClientID)
	authGuard, err := httpserver.NewAuthGuard(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("app: failed to init auth guard: %w", err)
	}

	// setting up the identity modules
	identityops := identity.Options{
		DB:  dbPool,
		Log: log,
		Cfg: identity.Config{
			Issuer: cfg.OAuthIssuer,

			GlobalSecret:      []byte(cfg.FositeGlobalSecret),
			JWTPrivateKeyPath: cfg.JWTPrivateKeyPath, // or
			AuthClientId:      authclientid,
			AuthClientSecret:  cfg.AuthClientSecret,
		},
		Renderer:   renderer,
		AuthGuard:  authGuard,
		PrivateKey: privateKey,
	}

	identityModule, err := identity.New(ctx, identityops)
	if err != nil {
		log.Fatal(ctx).Msg(fmt.Sprintf("app: failed to initialize identity module: %v", err))
		return nil, fmt.Errorf("app: failed to initialize identity module: %w", err)
	}

	// setting up the leadcrm module
	leadcrmops := leadcrm.Options{
		DB:       dbPool,
		Log:      log,
		Renderer: renderer,
	}
	leadCRMModule, err := leadcrm.New(ctx, leadcrmops)
	if err != nil {
		log.Fatal(ctx).Msg(fmt.Sprintf("app: failed to initialize leadcrm module: %v", err))
		return nil, fmt.Errorf("app: failed to initialize leadcrm module: %w", err)
	}

	// Register Modules
	if err := identityModule.Services.ModuleService.RegisterManifest(ctx, identityModule.Manifest); err != nil {
		log.Error(ctx).Err(err).Msg("app: failed to register identity manifest")
	}
	if err := identityModule.Services.ModuleService.RegisterManifest(ctx, leadCRMModule.Manifest); err != nil {
		log.Error(ctx).Err(err).Msg("app: failed to register leadcrm manifest")
	}

	// 3. Define the Global Router
	// This is where we will mount modules later (e.g., identity.RegisterRoutes(r))
	routerSetup := func(r chi.Router) {
		fileServer(r, "/assets", http.Dir("./assets"))
		r.Get("/", templ.Handler(pages.Landing()).ServeHTTP)
		
		// Register Module Routes directly on the main router
		identityModule.RegisterRoutes(r)
		leadCRMModule.RegisterRoutes(r)
	}

	// 4. Initialize HTTP Server
	srv := httpserver.NewServer(httpserver.Options{
		Addr: cfg.HTTPAddr,
	}, log, routerSetup)

	return &App{
		Config:   cfg,
		Logger:   log,
		Server:   srv,
		Identity: identityModule,
		LeadCRM:  leadCRMModule,
	}, nil
}

// Start runs the application and blocks until the context is cancelled
func (a *App) Start(ctx context.Context) error {
	return a.Server.Start(ctx)
}
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
