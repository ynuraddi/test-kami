package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type TxBeginner interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type transaction struct {
	conn TxBeginner
}

func NewTxManager(conn TxBeginner) *transaction {
	return &transaction{
		conn: conn,
	}
}

func (t *transaction) Execute(ctx context.Context, f func(txCtx context.Context) error, options pgx.TxOptions) (err error) {
	// WARNING: нужно использовать пул подключений
	tx, err := t.conn.BeginTx(ctx, options)
	if err != nil {
		return err
	}

	if err = f(injectTx(ctx, tx)); err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			return fmt.Errorf("%w | rollback err: %w", err, rollbackErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

type txKey struct{}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func extractTx(ctx context.Context) (tx pgx.Tx) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

func solveTx(conn DBTX, ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return conn
}
