package goerd

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/jmoiron/sqlx"
)

func WithTx(ctx context.Context, f func(context.Context) error) error {
	if HasTx(ctx) {
		return f(ctx)
	}
	ctxTx, err := withTx(ctx)
	if err != nil {
		return err
	}
	commit := false
	defer func() {
		if r := recover(); r != nil || !commit {
			if r != nil {
				log.Printf("!!! TRANSACTION PANIC !!! : %s\n%s", r, string(debug.Stack()))
			}
			if e := Rollback(ctxTx); e != nil {
				err = e
			} else if r != nil {
				err = fmt.Errorf("transaction panic: %s", r)
			}
		} else if commit {
			if e := Commit(ctxTx); e != nil {
				err = e
			}
		}
	}()

	if e := f(ctxTx); e != nil {
		return e
	}

	commit = true
	return nil
}

type CtxSqlxDb struct{}

func WithSqlxDb(ctx context.Context, d *sqlx.DB) context.Context {
	return context.WithValue(ctx, CtxSqlxDb{}, d)
}

func SqlxDbFromContext(ctx context.Context) *sqlx.DB {
	u, ok := ctx.Value(CtxSqlxDb{}).(*sqlx.DB)
	if u == nil || !ok {
		return nil
	}
	return u
}

type CtxSqlxTx struct{}

func WithSqlxTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, CtxSqlxTx{}, tx)
}

func SqlxTxFromContext(ctx context.Context) *sqlx.Tx {
	u, ok := ctx.Value(CtxSqlxTx{}).(*sqlx.Tx)
	if u == nil || !ok {
		return nil
	}
	return u
}

func withTx(ctx context.Context) (context.Context, error) {
	tx := SqlxTxFromContext(ctx)
	if tx != nil {
		return ctx, nil
	}
	d := SqlxDbFromContext(ctx)
	if d == nil {
		panic("sqlx DB not found in context")
	}
	tx, err := d.BeginTxx(ctx, nil)
	if err != nil {
		return ctx, err
	}
	return WithSqlxTx(ctx, tx), nil
}

func HasTx(ctx context.Context) bool {
	return SqlxTxFromContext(ctx) != nil
}

func Commit(ctx context.Context) error {
	tx := SqlxTxFromContext(ctx)
	if tx == nil {
		panic("transaction not found in context")
	}
	return tx.Commit()
}

func Rollback(ctx context.Context) error {
	tx := SqlxTxFromContext(ctx)
	if tx == nil {
		panic("nil transaction not found in context")
	}
	return tx.Rollback()
}
