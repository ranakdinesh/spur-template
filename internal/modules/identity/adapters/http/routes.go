package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/spurbase/spur/internal/modules/identity/adapters/http/handlers"
	"github.com/spurbase/spur/internal/platform/httpserver"
)

func RegisterRoutes(r chi.Router, clientHandler *handlers.ClientHandler, regHandler *handlers.RegistrationHandler, authHandler *handlers.AuthHandler, verificationGuard *httpserver.VerificationGuard, authGuard *httpserver.AuthGuard, apiKeyHandler *handlers.APIKeyHandler, rbacHandler *handlers.RBACHandler) {
	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", regHandler.RegisterTenant)
		r.Post("/login", authHandler.Login)

		r.Post("/password-reset/request", authHandler.RequestPasswordReset)
		r.Post("/password-reset/reset", authHandler.ResetPassword)

		r.Post("/magic-link/request", authHandler.RequestMagicLink)
		r.Get("/magic-link/login", authHandler.LoginWithMagicLink)
	})

	r.Route("/oauth2", func(r chi.Router) {
		r.Get("/auth", authHandler.Authorize)
		r.Post("/token", authHandler.Token)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		// Enforce Authentication
		r.Use(authGuard.ChiMiddleware)
		// Verification check enforcement
		r.Use(verificationGuard.ChiMiddleware)

		r.Get("/auth/me", authHandler.GetMe)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", regHandler.RegisterUser)
			r.Get("/", regHandler.ListUsers)
			r.Put("/{id}", regHandler.UpdateUser)
			r.Put("/{id}/lock", regHandler.UpdateUserLockStatus)
			r.Put("/{id}/password", regHandler.UpdateUserPassword)
			r.Delete("/{id}", regHandler.DeleteUser)
		})

		r.Route("/clients", func(r chi.Router) {
			r.Post("/", clientHandler.CreateClient)
			r.Get("/", clientHandler.ListClients)
			r.Get("/public", clientHandler.ListPublicClients)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", clientHandler.GetClient)
				r.Put("/", clientHandler.UpdateClient)
				r.Delete("/", clientHandler.DeleteClient)
			})
		})

		r.Route("/apikeys", func(r chi.Router) {
			r.Post("/", apiKeyHandler.CreateAPIKey)
			r.Get("/", apiKeyHandler.ListAPIKeys)
			r.Delete("/{id}", apiKeyHandler.DeleteAPIKey)
		})

		r.Route("/tenants", func(r chi.Router) {
			r.Post("/", rbacHandler.CreateTenant)
			r.Get("/", rbacHandler.ListTenants)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", rbacHandler.GetTenant)
				r.Put("/", rbacHandler.UpdateTenant)
				r.Delete("/", rbacHandler.DeleteTenant)
				r.Get("/modules", rbacHandler.ListTenantModules)
				r.Post("/modules", rbacHandler.EnableModuleForTenant)
				r.Delete("/modules", rbacHandler.DisableModuleForTenant)
			})
		})

		r.Route("/modules", func(r chi.Router) {
			r.Post("/", rbacHandler.CreateModule)
			r.Get("/", rbacHandler.ListModules)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", rbacHandler.GetModule)
				r.Put("/", rbacHandler.UpdateModule)
				r.Delete("/", rbacHandler.DeleteModule)
				r.Get("/permissions", rbacHandler.ListPermissionsByModule)
			})
		})

		r.Route("/permissions", func(r chi.Router) {
			r.Post("/", rbacHandler.CreatePermission)
			r.Get("/", rbacHandler.ListPermissions)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", rbacHandler.GetPermission)
				r.Put("/", rbacHandler.UpdatePermission)
				r.Delete("/", rbacHandler.DeletePermission)
			})
		})

		r.Route("/roles", func(r chi.Router) {
			r.Post("/", rbacHandler.CreateRole)
			r.Get("/", rbacHandler.ListRoles)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", rbacHandler.GetRole)
				r.Put("/", rbacHandler.UpdateRole)
				r.Delete("/", rbacHandler.DeleteRole)
			})
			r.Route("/{roleID}/permissions", func(r chi.Router) {
				r.Get("/", rbacHandler.ListRolePermissions)
				r.Post("/", rbacHandler.AssignPermissionToRole)
				r.Delete("/{permissionID}", rbacHandler.RemovePermissionFromRole)
			})
		})
	})
}
