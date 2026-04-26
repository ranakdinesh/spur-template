package services

import (
	"context"

	"net/http"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
)

type FositeService struct {
	clientRepo ports.ClientRepo
	provider   fosite.OAuth2Provider
}

func NewFositeService(clientRepo ports.ClientRepo, provider fosite.OAuth2Provider) *FositeService {
	return &FositeService{
		clientRepo: clientRepo,
		provider:   provider,
	}
}

func (s *FositeService) CreateClient(ctx context.Context, cmd ports.CreateClientCmd) (*sqlc.FositeClients, error) {
	// Generate new client ID if not provided? Fosite creates its own or we do.
	// The DB schema says `id TEXT PRIMARY KEY`.
	// Usually one generates a random ID.
	// Let's generate a UUID for ID if logic requires, but fosite clients usually have specific IDs.
	// For now let's assume the ID is generated here.
	clientID := uuid.New().String()

	params := sqlc.CreateClientParams{
		ID:            clientID,
		TenantID:      cmd.TenantID,
		ClientSecret:  cmd.ClientSecret,
		RedirectUris:  cmd.RedirectURIs,
		GrantTypes:    cmd.GrantTypes,
		ResponseTypes: cmd.ResponseTypes,
		Scopes:        cmd.Scopes,
		Audience:      cmd.Audience,
		Public:        cmd.Public,
		Active:        true,
	}

	client, err := s.clientRepo.CreateClient(ctx, params)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (s *FositeService) GetClient(ctx context.Context, id string) (*sqlc.FositeClients, error) {
	client, err := s.clientRepo.GetClient(ctx, id)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (s *FositeService) ListClients(ctx context.Context, tenantID uuid.UUID) ([]*sqlc.FositeClients, error) {
	clients, err := s.clientRepo.ListClients(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// Convert to pointers
	var result []*sqlc.FositeClients
	for _, c := range clients {
		// Use a local variable to take address
		val := c
		result = append(result, &val)
	}
	return result, nil
}

func (s *FositeService) ListPublicClients(ctx context.Context) ([]*sqlc.FositeClients, error) {
	clients, err := s.clientRepo.ListPublicClients(ctx)
	if err != nil {
		return nil, err
	}
	var result []*sqlc.FositeClients
	for _, c := range clients {
		val := c
		result = append(result, &val)
	}
	return result, nil
}

func (s *FositeService) UpdateClient(ctx context.Context, id string, cmd ports.UpdateClientCmd) error {
	// Update config
	configParams := sqlc.UpdateClientConfigParams{
		ID:           id,
		RedirectUris: cmd.RedirectURIs,
		Scopes:       cmd.Scopes,
		GrantTypes:   cmd.GrantTypes,
	}
	if err := s.clientRepo.UpdateClientConfig(ctx, configParams); err != nil {
		return err
	}

	// Update Secret if provided
	if cmd.ClientSecret != nil {
		secretParams := sqlc.UpdateClientSecretParams{
			ID:           id,
			ClientSecret: cmd.ClientSecret,
		}
		if err := s.clientRepo.UpdateClientSecret(ctx, secretParams); err != nil {
			return err
		}
	}

	// Update Status if provided
	if cmd.Active != nil {
		statusParams := sqlc.ToggleClientStatusParams{
			ID:     id,
			Active: *cmd.Active,
		}
		if err := s.clientRepo.ToggleClientStatus(ctx, statusParams); err != nil {
			return err
		}
	}

	return nil
}

func (s *FositeService) DeleteClient(ctx context.Context, id string) error {
	return s.clientRepo.DeleteClient(ctx, id)
}

// CitualSession wraps openid.DefaultSession to implement JWTSessionContainer
type CitualSession struct {
	*openid.DefaultSession
}

// Interface Guard to ensure CitualSession correctly implements JWTSessionContainer
var _ oauth2.JWTSessionContainer = (*CitualSession)(nil)

func (s *CitualSession) GetJWTClaims() jwt.JWTClaimsContainer {
	claims := &jwt.JWTClaims{
		Subject:   s.Subject,
		Issuer:    "", // Will be set by defaults
		Audience:  s.DefaultSession.Claims.Audience,
		JTI:       s.DefaultSession.Claims.JTI,
		IssuedAt:  s.DefaultSession.Claims.IssuedAt,
		ExpiresAt: s.DefaultSession.Claims.ExpiresAt,
		Extra:     make(map[string]interface{}),
	}

	// Copy extra claims
	if s.DefaultSession.Claims != nil && s.DefaultSession.Claims.Extra != nil {
		for k, v := range s.DefaultSession.Claims.Extra {
			claims.Extra[k] = v
		}
	}
	claims.Extra["username"] = s.Username

	return claims
}

func (s *CitualSession) GetJWTHeader() *jwt.Headers {
	return &jwt.Headers{
		Extra: map[string]interface{}{
			"typ": "JWT",
			"alg": "RS256",
		},
	}
}

func (s *CitualSession) Clone() fosite.Session {
	if s == nil {
		return nil
	}
	return &CitualSession{
		DefaultSession: s.DefaultSession.Clone().(*openid.DefaultSession),
	}
}

// OAuth2 Handlers Implementation

func (s *FositeService) NewAuthorizeRequest(ctx context.Context, r *http.Request) (fosite.AuthorizeRequester, error) {
	return s.provider.NewAuthorizeRequest(ctx, r)
}

func (s *FositeService) NewAuthorizeResponse(ctx context.Context, ar fosite.AuthorizeRequester, session *ports.SessionUserData) (fosite.AuthorizeResponder, error) {
	// Use CitualSession which satisfies JWTSessionContainer AND OpenID Session
	sess := &CitualSession{
		DefaultSession: &openid.DefaultSession{
			Claims: &jwt.IDTokenClaims{
				Subject: session.UserID,
				Extra:   make(map[string]interface{}),
			},
			Headers:  &jwt.Headers{},
			Subject:  session.UserID,
			Username: session.UserID,
		},
	}

	// Set required claims for Citual
	sess.Claims.Extra["tid"] = session.TenantID
	sess.Claims.Extra["sa"] = session.IsSuperAdmin
	sess.Claims.Extra["av"] = session.AuthzVersion
	sess.Claims.Extra["roles"] = session.Roles

	// Compatibility fallback
	sess.Claims.Extra["tenant_id"] = session.TenantID

	return s.provider.NewAuthorizeResponse(ctx, ar, sess)
}

func (s *FositeService) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar fosite.AuthorizeRequester, resp fosite.AuthorizeResponder) {
	s.provider.WriteAuthorizeResponse(ctx, rw, ar, resp)
}

func (s *FositeService) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar fosite.AuthorizeRequester, err error) {
	s.provider.WriteAuthorizeError(ctx, rw, ar, err)
}

func (s *FositeService) NewAccessRequest(ctx context.Context, r *http.Request) (fosite.AccessRequester, error) {
	// Session factory
	sess := &CitualSession{
		DefaultSession: &openid.DefaultSession{
			Claims:  &jwt.IDTokenClaims{Extra: make(map[string]interface{})},
			Headers: &jwt.Headers{},
		},
	}
	return s.provider.NewAccessRequest(ctx, r, sess)
}

func (s *FositeService) NewAccessResponse(ctx context.Context, ar fosite.AccessRequester) (fosite.AccessResponder, error) {
	return s.provider.NewAccessResponse(ctx, ar)
}

func (s *FositeService) WriteAccessResponse(ctx context.Context, rw http.ResponseWriter, ar fosite.AccessRequester, resp fosite.AccessResponder) {
	s.provider.WriteAccessResponse(ctx, rw, ar, resp)
}

func (s *FositeService) WriteAccessError(ctx context.Context, rw http.ResponseWriter, ar fosite.AccessRequester, err error) {
	s.provider.WriteAccessError(ctx, rw, ar, err)
}
