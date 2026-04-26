package postgres

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/adapters/postgres/sqlc"
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Adapter struct {
	db *sql.DB
	q  *sqlc.Queries
}

func New(pool *pgxpool.Pool) *Adapter {
	db := stdlib.OpenDBFromPool(pool)
	return &Adapter{
		db: db,
		q:  sqlc.New(db),
	}
}

func (a *Adapter) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// Stub implementation of transaction
	// Real impl: tx, _ := a.db.BeginTx(ctx, ...); ctxWithTx := context.WithValue(ctx, txKey, tx); err := fn(ctxWithTx); if err ... commit else rollback
	return fn(ctx)
}
