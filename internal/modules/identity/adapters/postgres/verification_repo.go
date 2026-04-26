package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
)

type VerificationRepo struct {
	store *Store
}

func NewVerificationRepo(store *Store) ports.VerificationRepo {
	return &VerificationRepo{store: store}
}

func (r *VerificationRepo) CreateChallenge(ctx context.Context, challenge *domain.VerificationChallenge) error {
	q := r.store.getQueries(ctx)

	// Convert domain.VerificationKind to string
	kindStr := string(challenge.Kind)

	_, err := q.CreateVerificationChallenge(ctx, sqlc.CreateVerificationChallengeParams{
		ID:        challenge.ID,
		TenantID:  challenge.TenantID,
		UserID:    challenge.UserID,
		Kind:      kindStr,
		TokenHash: challenge.TokenHash,
		ExpiresAt: challenge.ExpiresAt,
	})
	return err
}

func (r *VerificationRepo) GetChallenge(ctx context.Context, userID uuid.UUID, kind domain.VerificationKind) (*domain.VerificationChallenge, error) {
	q := r.store.getQueries(ctx)

	row, err := q.GetVerificationChallenge(ctx, sqlc.GetVerificationChallengeParams{
		UserID: userID,
		Kind:   string(kind),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return mapChallengeToDomain(row), nil
}

func (r *VerificationRepo) GetChallengeByToken(ctx context.Context, token string, kind domain.VerificationKind) (*domain.VerificationChallenge, error) {
	q := r.store.getQueries(ctx)

	row, err := q.GetVerificationChallengeByToken(ctx, sqlc.GetVerificationChallengeByTokenParams{
		TokenHash: token,
		Kind:      string(kind),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return mapChallengeToDomain(row), nil
}

func (r *VerificationRepo) MarkChallengeConsumed(ctx context.Context, id uuid.UUID) error {
	q := r.store.getQueries(ctx)
	return q.MarkChallengeConsumed(ctx, id)
}

func (r *VerificationRepo) DeleteExpiredChallenges(ctx context.Context) error {
	q := r.store.getQueries(ctx)
	return q.DeleteExpiredChallenges(ctx)
}

func mapChallengeToDomain(row sqlc.VerificationChallenges) *domain.VerificationChallenge {
	challenge := &domain.VerificationChallenge{
		ID:        row.ID,
		TenantID:  row.TenantID,
		UserID:    row.UserID,
		Kind:      domain.VerificationKind(row.Kind),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}

	if row.ConsumedAt.Valid {
		val := row.ConsumedAt.Time
		challenge.ConsumedAt = &val
	}

	return challenge
}
