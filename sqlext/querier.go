package sqlext

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Querier interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	Rebind(query string) string
}

type txKey struct{}

func ContextWithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func GetQuerier(ctx context.Context, db *sqlx.DB) Querier {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok && tx != nil {
		return tx
	}
	return db
}
