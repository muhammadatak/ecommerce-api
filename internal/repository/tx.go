package repository

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/muhammadatak/go-ecommerce-api/domain"
)

type txContextKey struct{}

type dbExecutor interface {
	From(from ...interface{}) *goqu.SelectDataset
	Insert(table interface{}) *goqu.InsertDataset
	Update(table interface{}) *goqu.UpdateDataset
	Delete(table interface{}) *goqu.DeleteDataset
}

type txRepository struct {
	db *goqu.Database
}

func NewTransactor(con *sql.DB) domain.TxRepository {
	return &txRepository{
		db: goqu.New("default", con),
	}
}

// WithTransaction implements [domain.Transactor].
func (t *txRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return withTransaction(ctx, t.db, fn)
}

func getExecutor(ctx context.Context, db *goqu.Database) dbExecutor {
	if tx, ok := ctx.Value(txContextKey{}).(*goqu.TxDatabase); ok && tx != nil {
		return tx
	}
	return db
}

func isTx(ctx context.Context) bool {
	tx, ok := ctx.Value(txContextKey{}).(*goqu.TxDatabase)
	return ok && tx != nil
}

func withTransaction(ctx context.Context, db *goqu.Database, fn func(ctx context.Context) error) error {
	if isTx(ctx) {
		return fn(ctx)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	txCtx := context.WithValue(ctx, txContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
