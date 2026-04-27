package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ranakdinesh/spur/internal/config"
	"github.com/ranakdinesh/spur/internal/infrastructure"
	"github.com/ranakdinesh/spur/internal/logger"
	// SPUR:IMPORTS:END
)

type App struct {
	Infra *infrastructure.Infra
	// SPUR:APP_FIELDS
	// SPUR:APP_FIELDS:END
}

func New(ctx context.Context) (*App, error) {
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	log := logger.NewWithOptions(logger.Options{
		Environment: cfg.AppEnv,
	})

	infra, err := infrastructure.Bootstrap(ctx, &cfg, log)
	if err != nil {
		return nil, err
	}

	// SPUR:MODULES
	// SPUR:MODULES:END

	infra.HTTP.Mount(func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		})
		// SPUR:ROUTES
		// SPUR:ROUTES:END
	})

	return &App{
		Infra: infra,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	return a.Infra.HTTP.Start(ctx)
}
