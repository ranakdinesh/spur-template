package ports

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
)

// Legacy Fosite Service (Keep as is for now or refactor later if time permits)
// For now, FositeService uses sqlc types because it was implemented first.
// Ideally it should also be domain-based. But let's focus on the new Registration flow first.
// BUT, mixing them is confusing.
// Let's refactor FositeService ports too if possible, but maybe minimal touch for now.
// Actually, `sqlc.FositeClients` is effectively a domain object for that sub-domain.
// Let's leave FositeService as is for this step to reduce risk, unless it breaks import cycles.

type ClientRepo interface {
	CreateClient(ctx context.Context, arg sqlc.CreateClientParams) (sqlc.FositeClients, error)
	GetClient(ctx context.Context, id string) (sqlc.FositeClients, error)
	GetActiveClient(ctx context.Context, id string) (sqlc.FositeClients, error)
	ListClients(ctx context.Context, tenantID uuid.UUID) ([]sqlc.FositeClients, error)
	ListPublicClients(ctx context.Context) ([]sqlc.FositeClients, error)
	UpdateClientSecret(ctx context.Context, arg sqlc.UpdateClientSecretParams) error
	ToggleClientStatus(ctx context.Context, arg sqlc.ToggleClientStatusParams) error
	UpdateClientConfig(ctx context.Context, arg sqlc.UpdateClientConfigParams) error
	DeleteClient(ctx context.Context, id string) error
}

type FositeSessionRepo interface {
	CreateSession(ctx context.Context, arg sqlc.CreateSessionParams) error
	GetSession(ctx context.Context, arg sqlc.GetSessionParams) (sqlc.FositeSessions, error)
	DeleteSessionByType(ctx context.Context, arg sqlc.DeleteSessionByTypeParams) error
	RevokeSessionByRequestId(ctx context.Context, requestID string) error
	RevokeSessionByRequestIdAndType(ctx context.Context, arg sqlc.RevokeSessionByRequestIdAndTypeParams) error
}

type FositeService interface {
	CreateClient(ctx context.Context, cmd CreateClientCmd) (*sqlc.FositeClients, error)
	GetClient(ctx context.Context, id string) (*sqlc.FositeClients, error)
	ListClients(ctx context.Context, tenantID uuid.UUID) ([]*sqlc.FositeClients, error)
	UpdateClient(ctx context.Context, id string, cmd UpdateClientCmd) error
	DeleteClient(ctx context.Context, id string) error
	ListPublicClients(ctx context.Context) ([]*sqlc.FositeClients, error)

	// OAuth2 Handlers
	NewAuthorizeRequest(ctx context.Context, r *http.Request) (fosite.AuthorizeRequester, error)
	NewAuthorizeResponse(ctx context.Context, ar fosite.AuthorizeRequester, session *SessionUserData) (fosite.AuthorizeResponder, error)
	WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar fosite.AuthorizeRequester, resp fosite.AuthorizeResponder)
	WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar fosite.AuthorizeRequester, err error)

	NewAccessRequest(ctx context.Context, r *http.Request) (fosite.AccessRequester, error)
	NewAccessResponse(ctx context.Context, ar fosite.AccessRequester) (fosite.AccessResponder, error)
	WriteAccessResponse(ctx context.Context, rw http.ResponseWriter, ar fosite.AccessRequester, resp fosite.AccessResponder)
	WriteAccessError(ctx context.Context, rw http.ResponseWriter, ar fosite.AccessRequester, err error)
}

type CreateClientCmd struct {
	TenantID      uuid.UUID
	ClientSecret  *string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	Scopes        []string
	Audience      []string
	Public        bool
}

type UpdateClientCmd struct {
	ClientSecret  *string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	Scopes        []string
	Audience      []string
	Active        *bool
}

// New Repositories (Registration Flow) - USING DOMAIN TYPES

type TenantRepo interface {
	CreateTenant(ctx context.Context, tenant *domain.Tenant) (*domain.Tenant, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	ListTenants(ctx context.Context) ([]*domain.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *domain.Tenant) error
	DeleteTenant(ctx context.Context, id uuid.UUID) error
}

type UserRepo interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUser(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByMobile(ctx context.Context, mobile string) (*domain.User, error)
	ListUsers(ctx context.Context, tenantID uuid.UUID) ([]*domain.User, error)
	ListAllUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	UpdateUserLockStatus(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, isLocked bool) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, passwordHash string) error
	DeleteUser(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	DeleteUserByEmail(ctx context.Context, email string) error
	DeleteUserByMobile(ctx context.Context, mobile string) error
}

type RoleRepo interface {
	CreateRole(ctx context.Context, role *domain.Role) (*domain.Role, error)
	GetRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	GetRoleByCode(ctx context.Context, code string, tenantID uuid.UUID) (*domain.Role, error)
	GetSystemRoleByCode(ctx context.Context, code string) (*domain.Role, error)
	ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error)
	ListAllRoles(ctx context.Context) ([]*domain.Role, error)
	UpdateRole(ctx context.Context, role *domain.Role) error
	DeleteRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// These might remain specific assignments, but ideally domain concepts:
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	ListUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)
}

type VerificationRepo interface {
	CreateChallenge(ctx context.Context, challenge *domain.VerificationChallenge) error
	GetChallenge(ctx context.Context, userID uuid.UUID, kind domain.VerificationKind) (*domain.VerificationChallenge, error)
	GetChallengeByToken(ctx context.Context, token string, kind domain.VerificationKind) (*domain.VerificationChallenge, error)
	MarkChallengeConsumed(ctx context.Context, id uuid.UUID) error
	DeleteExpiredChallenges(ctx context.Context) error
}

type APIKeyRepo interface {
	CreateAPIKey(ctx context.Context, key *domain.APIKey) error
	GetAPIKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.APIKey, error)
	GetAPIKeyByPrefix(ctx context.Context, prefix string) (*domain.APIKey, error)
	ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error)
	UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error
	DeleteAPIKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

type ModuleRepo interface {
	CreateModule(ctx context.Context, m *domain.Module) (*domain.Module, error)
	UpsertModule(ctx context.Context, m *domain.Module) (*domain.Module, error)
	GetModule(ctx context.Context, id uuid.UUID) (*domain.Module, error)
	ListModules(ctx context.Context) ([]*domain.Module, error)
	UpdateModule(ctx context.Context, m *domain.Module) error
	DeleteModule(ctx context.Context, id uuid.UUID) error

	EnableModuleForTenant(ctx context.Context, tenantID, moduleID uuid.UUID) error
	DisableModuleForTenant(ctx context.Context, tenantID, moduleID uuid.UUID) error
	ListTenantModules(ctx context.Context, tenantID uuid.UUID) ([]*domain.Module, error)
}

type PermissionRepo interface {
	CreatePermission(ctx context.Context, p *domain.Permission) (*domain.Permission, error)
	UpsertPermission(ctx context.Context, p *domain.Permission) error
	GetPermission(ctx context.Context, id uuid.UUID) (*domain.Permission, error)
	ListPermissions(ctx context.Context) ([]*domain.Permission, error)
	ListPermissionsByModule(ctx context.Context, moduleID uuid.UUID) ([]*domain.Permission, error)
	UpdatePermission(ctx context.Context, p *domain.Permission) error
	DeletePermission(ctx context.Context, id uuid.UUID) error
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error)
}

// Registration Service

type RegistrationService interface {
	RegisterTenant(ctx context.Context, cmd RegisterTenantCmd) (*RegisteredTenantResult, error)
	CreateUser(ctx context.Context, cmd CreateUserCmd) (*domain.User, error)
	ListUsers(ctx context.Context, tenantID uuid.UUID) ([]*domain.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, cmd UpdateUserCmd) error
	UpdateUserLockStatus(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, isLocked bool) error
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, newPassword string) error
	DeleteUser(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) error
}

type UpdateUserCmd struct {
	FirstName string
	LastName  string
	Mobile    string
	Roles     []string
}

type RegisterTenantCmd struct {
	FirstName   string      `json:"first_name"`
	LastName    string      `json:"last_name"`
	CompanyName string      `json:"company_name"`
	Email       string      `json:"email"`
	Mobile      string      `json:"mobile"`
	Password    string      `json:"password"`
	ModuleIDs   []uuid.UUID `json:"module_ids"`
	AutoVerify  bool        `json:"auto_verify"`
}

type RegisteredTenantResult struct {
	Tenant *domain.Tenant
	User   *domain.User
}

type CreateUserCmd struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	Mobile       string    `json:"mobile"`
	Password     string    `json:"password"`
	Roles        []string  `json:"roles"`
	IsSuperAdmin bool      `json:"is_super_admin"`
}

type LoginCmd struct {
	Identifier string // Email or Mobile
	Email      string // Deprecated: Use Identifier
	Password   string
	IPAddress  string
	UserAgent  string
}

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	TenantID  uuid.UUID
}

type SessionUserData struct {
	UserID       string
	TenantID     string
	IsSuperAdmin bool
	AuthzVersion int
	Roles        []string
}

type AuthService interface {
	Login(ctx context.Context, cmd LoginCmd) (*Session, error)
	RequestOTP(ctx context.Context, input RequestOTPInput) error
	ResendOTP(ctx context.Context, input RequestOTPInput) error
	LoginWithOTP(ctx context.Context, input LoginWithOTPInput) (*Session, error)
	GetVerificationStatus(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (*VerificationStatus, error)

	// Password Recovery
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error

	// Magic Links
	RequestMagicLink(ctx context.Context, identifier string) error
	LoginWithMagicLink(ctx context.Context, token string) (*Session, error)

	GetSession(ctx context.Context, token string) (*Session, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (*CurrentUser, error)
	Logout(ctx context.Context, token string) error
}

type CurrentUser struct {
	ID           uuid.UUID `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	TenantID     uuid.UUID `json:"tenant_id"`
	IsSuperAdmin bool      `json:"is_super_admin"`
	Roles        []string  `json:"roles"`
	Permissions  []string  `json:"permissions"`
}

type RequestOTPInput struct {
	Identifier string // Email or Mobile
	Channel    string // "email" or "mobile"
}

type VerificationStatus struct {
	IsVerified         bool
	GracePeriodExpired bool
}

type LoginWithOTPInput struct {
	Identifier string
	Code       string
	IPAddress  string
	UserAgent  string
}

type SessionRepo interface {
	CreateUserSession(ctx context.Context, s *Session) error
	GetUserSessionByToken(ctx context.Context, token string) (*Session, error)
	DeleteUserSession(ctx context.Context, token string) error
	DeleteUserSessionsByUserID(ctx context.Context, userID uuid.UUID) error
}

// TxManager abstracts database transactions so services don't depend on *postgres.Store.
type TxManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// APIKeyService interface for the API key module.
type APIKeyService interface {
	CreateAPIKey(ctx context.Context, input CreateAPIKeyCmd) (*CreateAPIKeyResult, error)
	Authenticate(ctx context.Context, fullKey string, origin string) (*domain.APIKey, error)
	ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error)
	DeleteAPIKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

// RBACService interface for role/permission management.
type RBACService interface {
	// Tenants
	CreateTenant(ctx context.Context, t *domain.Tenant) (*domain.Tenant, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	ListTenants(ctx context.Context) ([]*domain.Tenant, error)
	UpdateTenant(ctx context.Context, t *domain.Tenant) error
	DeleteTenant(ctx context.Context, id uuid.UUID) error
	// Roles
	CreateRole(ctx context.Context, r *domain.Role) (*domain.Role, error)
	GetRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error)
	ListAllRoles(ctx context.Context) ([]*domain.Role, error)
	UpdateRole(ctx context.Context, r *domain.Role) error
	DeleteRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	// Permissions
	CreatePermission(ctx context.Context, p *domain.Permission) (*domain.Permission, error)
	GetPermission(ctx context.Context, id uuid.UUID) (*domain.Permission, error)
	ListPermissions(ctx context.Context) ([]*domain.Permission, error)
	ListPermissionsByModule(ctx context.Context, moduleID uuid.UUID) ([]*domain.Permission, error)
	UpdatePermission(ctx context.Context, p *domain.Permission) error
	DeletePermission(ctx context.Context, id uuid.UUID) error
	AssignPermissionToRole(ctx context.Context, roleID, permID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permID uuid.UUID) error
	ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error)
	// Modules
	CreateModule(ctx context.Context, m *domain.Module) (*domain.Module, error)
	GetModule(ctx context.Context, id uuid.UUID) (*domain.Module, error)
	ListModules(ctx context.Context) ([]*domain.Module, error)
	UpdateModule(ctx context.Context, m *domain.Module) error
	DeleteModule(ctx context.Context, id uuid.UUID) error
	EnableModuleForTenant(ctx context.Context, tenantID, moduleID uuid.UUID) error
	DisableModuleForTenant(ctx context.Context, tenantID, moduleID uuid.UUID) error
	ListTenantModules(ctx context.Context, tenantID uuid.UUID) ([]*domain.Module, error)
}

// ─── APIKey command/result types ─────────────────────────────────────────────

type CreateAPIKeyCmd struct {
	TenantID       uuid.UUID
	Name           string
	Type           string   // "secret" | "publishable"
	Scopes         []string
	AllowedOrigins []string
}

type CreateAPIKeyResult struct {
	Key    *domain.APIKey
	Secret string // returned once only at creation
}
