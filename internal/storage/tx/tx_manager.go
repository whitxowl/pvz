package tx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whitxowl/pvz.git/pkg/postgres"
)

type txManager struct {
	db *pgxpool.Pool
}

func NewTxManager(db *pgxpool.Pool) *txManager {
	return &txManager{db: db}
}

type txKey struct{}

func (m *txManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, txKey{}, tx)

	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func TxFromCtx(ctx context.Context, db postgres.DB) postgres.DB {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return db
}
