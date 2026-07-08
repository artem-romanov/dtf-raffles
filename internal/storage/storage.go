package storage

import (
	"context"
	"database/sql"
	"dtf/game_draw/internal/domain"
)

type txKeyT struct{}

var txKey = txKeyT{}
var _ domain.Transactor = (*SqlTransactor)(nil)

type Provider struct {
	db *sql.DB
}

func NewProvider(db *sql.DB) *Provider {
	return &Provider{db: db}
}

func (p *Provider) Ext(ctx context.Context) domain.DBTX {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}
	return p.db
}

type SqlTransactor struct {
	db *sql.DB
}

func NewSqlTransactor(db *sql.DB) *SqlTransactor {
	return &SqlTransactor{db: db}
}

func (st *SqlTransactor) WithTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	return st.WithTransactionOptions(ctx, nil, fn)
}

func (st *SqlTransactor) WithTransactionOptions(
	ctx context.Context,
	opts *sql.TxOptions,
	fn func(ctx context.Context) error,
) error {
	if _, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return fn(ctx)
	}

	tx, err := st.db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	txCtx := context.WithValue(ctx, txKey, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit()
}
