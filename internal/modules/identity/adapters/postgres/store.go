package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spurbase/spur/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/spurbase/spur/internal/platform/db"
)

type Store struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}
func (s *Store) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return db.RunInTx(ctx, s.Pool, fn)
}

func (s *Store) getQueries(ctx context.Context) *sqlc.Queries {
	// MODIFIED: Use the platform helper to find the transaction
	if tx := db.GetTx(ctx); tx != nil {
		return s.Queries.WithTx(tx)
	}

	// Fallback to pool (Auto-commit)
	return s.Queries
}
