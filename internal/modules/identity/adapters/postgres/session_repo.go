package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
)

func (s *Store) CreateUserSession(ctx context.Context, session *ports.Session) error {
	arg := sqlc.CreateUserSessionParams{
		ID:        session.ID,
		UserID:    session.UserID,
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
		IpAddress: nil,
		UserAgent: nil,
	}
	_, err := s.getQueries(ctx).CreateUserSession(ctx, arg)
	return err
}

func (s *Store) GetUserSessionByToken(ctx context.Context, token string) (*ports.Session, error) {
	row, err := s.getQueries(ctx).GetUserSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &ports.Session{
		ID:        row.ID,
		UserID:    row.UserID,
		Token:     row.Token,
		ExpiresAt: row.ExpiresAt, // direct access
		TenantID:  row.TenantID,
	}, nil
}

func (s *Store) DeleteUserSession(ctx context.Context, token string) error {
	return s.getQueries(ctx).DeleteUserSession(ctx, token)
}

func (s *Store) DeleteUserSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.getQueries(ctx).DeleteUserSessionsByUserID(ctx, userID)
}
