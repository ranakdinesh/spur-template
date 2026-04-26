package ports

import "context"

type TransactionManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
