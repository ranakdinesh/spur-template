package postgres

import (
	"context"

	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
)

// CreateSession creates a new Fosite session
// CreateSession creates a new Fosite session
func (s *Store) CreateSession(ctx context.Context, arg sqlc.CreateSessionParams) error {
	return s.getQueries(ctx).CreateSession(ctx, arg)
}

// GetSession retrieves a Fosite session
func (s *Store) GetSession(ctx context.Context, arg sqlc.GetSessionParams) (sqlc.FositeSessions, error) {
	return s.getQueries(ctx).GetSession(ctx, arg)
}

// DeleteSessionByType deletes sessions of a specific type (e.g., access_token) for a signature
func (s *Store) DeleteSessionByType(ctx context.Context, arg sqlc.DeleteSessionByTypeParams) error {
	_, err := s.getQueries(ctx).DeleteSessionByType(ctx, arg)
	return err
}

// RevokeSessionByRequestId revokes all sessions associated with a request ID
func (s *Store) RevokeSessionByRequestId(ctx context.Context, requestID string) error {
	return s.getQueries(ctx).RevokeSessionByRequestId(ctx, requestID)
}

// RevokeSessionByRequestIdAndType revokes sessions of a specific type for a request ID
func (s *Store) RevokeSessionByRequestIdAndType(ctx context.Context, arg sqlc.RevokeSessionByRequestIdAndTypeParams) error {
	return s.getQueries(ctx).RevokeSessionByRequestIdAndType(ctx, arg)
}
