// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package fosite_store

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ory/fosite"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/platform/httpserver"
)

var (
	_ fosite.Storage = (*FositeStore)(nil)
)

// FositeStore is a fosite.Storage implementation that uses a PostgreSQL database.
type FositeStore struct {
	store *postgres.Store
}

// NewStore creates a new FositeStore.
func NewStore(store *postgres.Store) *FositeStore {
	return &FositeStore{store: store}
}

func (s *FositeStore) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	client, err := s.store.GetClient(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	var secret []byte
	if client.ClientSecret != nil {
		secret = []byte(*client.ClientSecret)
	}

	return &fosite.DefaultClient{
		ID:            client.ID,
		Secret:        secret,
		RedirectURIs:  client.RedirectUris,
		GrantTypes:    client.GrantTypes,
		ResponseTypes: client.ResponseTypes,
		Audience:      client.Audience,
		Scopes:        client.Scopes,
		Public:        client.Public,
	}, nil
}

func (s *FositeStore) createSession(ctx context.Context, signature string, requester fosite.Requester, sessionType string) error {
	sessionData, err := json.Marshal(requester.GetSession())
	if err != nil {
		return fosite.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	tenantIDStr := httpserver.GetTenantID(ctx)
	if tenantIDStr == "" {
		// Fallback: try to get from session extra data
		if sess, ok := requester.GetSession().(*fosite.DefaultSession); ok {
			if tid, ok := sess.Extra["tenant_id"].(string); ok {
				tenantIDStr = tid
			}
		}
	}

	if tenantIDStr == "" {
		return fosite.ErrServerError.WithDebug("Tenant ID not found in context")
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fosite.ErrServerError.WithWrap(err).WithDebug("Invalid Tenant ID in context")
	}

	formDataBytes, err := json.Marshal(requester.GetRequestForm())
	if err != nil {
		return fosite.ErrServerError.WithWrap(err).WithDebug("Failed to marshal form data")
	}

	params := sqlc.CreateSessionParams{
		Signature:   signature,
		RequestID:   requester.GetID(),
		ClientID:    requester.GetClient().GetID(),
		TenantID:    tenantID,
		Subject:     requester.GetSession().GetSubject(),
		Type:        sessionType,
		Active:      true,
		RequestedAt: requester.GetRequestedAt(),
		ExpiresAt:   requester.GetSession().GetExpiresAt(fosite.TokenType(sessionType)),
		FormData:    formDataBytes,
		SessionData: sessionData,
	}

	err = s.store.CreateSession(ctx, params)
	if err != nil {
		return fosite.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	return nil
}

func (s *FositeStore) getSession(ctx context.Context, signature string, session fosite.Session, sessionType string) (fosite.Requester, error) {
	dbSession, err := s.store.GetSession(ctx, sqlc.GetSessionParams{
		Signature: signature,
		Type:      sessionType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	if !dbSession.Active {
		return nil, fosite.ErrInactiveToken.WithDebug("Token is inactive")
	}

	if session != nil {
		if err := json.Unmarshal(dbSession.SessionData, session); err != nil {
			return nil, fosite.ErrServerError.WithWrap(err)
		}
	}

	var form url.Values
	if err := json.Unmarshal(dbSession.FormData, &form); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	client, err := s.GetClient(ctx, dbSession.ClientID)
	if err != nil {
		return nil, err
	}

	return &fosite.Request{
		ID:          dbSession.RequestID,
		RequestedAt: dbSession.RequestedAt,
		Client:      client,
		Session:     session,
		Form:        form,
	}, nil
}

func (s *FositeStore) deleteSession(ctx context.Context, signature string, sessionType string) error {
	return s.store.DeleteSessionByType(ctx, sqlc.DeleteSessionByTypeParams{
		Signature: signature,
		Type:      sessionType,
	})
}

func (s *FositeStore) CreateAuthorizeCodeSession(ctx context.Context, signature string, requester fosite.Requester) error {
	return s.createSession(ctx, signature, requester, "authorization_code")
}

func (s *FositeStore) GetAuthorizeCodeSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return s.getSession(ctx, signature, session, "authorization_code")
}

func (s *FositeStore) InvalidateAuthorizeCodeSession(ctx context.Context, signature string) error {
	// Delete the authorization code row by its signature to prevent reuse
	return s.deleteSession(ctx, signature, "authorization_code")
}

func (s *FositeStore) CreatePKCERequestSession(ctx context.Context, signature string, requester fosite.Requester) error {
	return s.createSession(ctx, signature, requester, "pkce_request")
}

func (s *FositeStore) GetPKCERequestSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return s.getSession(ctx, signature, session, "pkce_request")
}

func (s *FositeStore) DeletePKCERequestSession(ctx context.Context, signature string) error {
	return s.deleteSession(ctx, signature, "pkce_request")
}

func (s *FositeStore) CreateAccessTokenSession(ctx context.Context, signature string, requester fosite.Requester) error {
	return s.createSession(ctx, signature, requester, "access_token")
}

func (s *FositeStore) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return s.getSession(ctx, signature, session, "access_token")
}

func (s *FositeStore) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	return s.deleteSession(ctx, signature, "access_token")
}

func (s *FositeStore) CreateRefreshTokenSession(ctx context.Context, signature string, accessSignature string, requester fosite.Requester) error {
	return s.createSession(ctx, signature, requester, "refresh_token")
}

func (s *FositeStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return s.getSession(ctx, signature, session, "refresh_token")
}

func (s *FositeStore) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	return s.deleteSession(ctx, signature, "refresh_token")
}

func (s *FositeStore) CreateOpenIDConnectSession(ctx context.Context, signature string, requester fosite.Requester) error {
	return s.createSession(ctx, signature, requester, "openid_connect")
}

func (s *FositeStore) GetOpenIDConnectSession(ctx context.Context, signature string, requester fosite.Requester) (fosite.Requester, error) {
	return s.getSession(ctx, signature, requester.GetSession(), "openid_connect")
}

func (s *FositeStore) DeleteOpenIDConnectSession(ctx context.Context, signature string) error {
	return s.deleteSession(ctx, signature, "openid_connect")
}

func (s *FositeStore) CreateTokenRevocation(ctx context.Context, signature string) error {
	return nil
}

func (s *FositeStore) GetTokenRevocation(ctx context.Context, signature string) error {
	return nil
}

func (s *FositeStore) DeleteTokenRevocation(ctx context.Context, signature string) error {
	return nil
}

func (s *FositeStore) RotateRefreshToken(ctx context.Context, requestID string, tokenSignature string) error {
	return s.RevokeRefreshToken(ctx, requestID)
}

func (s *FositeStore) RevokeRefreshToken(ctx context.Context, requestID string) error {
	return s.store.RevokeSessionByRequestIdAndType(ctx, sqlc.RevokeSessionByRequestIdAndTypeParams{
		RequestID: requestID,
		Type:      "refresh_token",
	})
}

func (s *FositeStore) RevokeAccessToken(ctx context.Context, requestID string) error {
	return s.store.RevokeSessionByRequestIdAndType(ctx, sqlc.RevokeSessionByRequestIdAndTypeParams{
		RequestID: requestID,
		Type:      "access_token",
	})
}

func (s *FositeStore) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	return nil
}

func (s *FositeStore) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	return nil
}
