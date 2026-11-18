package repository

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/muhammadatak/go-ecommerce-api/domain"
)

type orderRepository struct {
	db *goqu.Database
}

func NewOrder(con *sql.DB) domain.OrderRepository {
	return &orderRepository{
		db: goqu.New("default", con),
	}
}
func (o *orderRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return withTransaction(ctx, o.db, fn)
}

// FindById implements [domain.OrderRepository].
func (o *orderRepository) FindById(ctx context.Context, id string) (domain.Order, error) {
	panic("unimplemented")
}

// Save implements [domain.OrderRepository].
func (o *orderRepository) Save(ctx context.Context, order domain.Order) error {
	executor := getExecutor(ctx, o.db).Insert("orders").Rows(order).Executor()
	_, err := executor.ExecContext(ctx)

	return err
}

func (o *orderRepository) Update(ctx context.Context, order *domain.Order) error {
	executor := getExecutor(ctx, o.db).Update("orders").Where(goqu.C("id").Eq(order.Id)).Set(goqu.Record{"total_price": order.TotalPrice}).Executor()
	_, err := executor.ExecContext(ctx)

	return err
}
