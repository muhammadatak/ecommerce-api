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

// FindAllByUserId implements [domain.OrderRepository].

// FindByAll implements [domain.OrderRepository].

func NewOrder(con *sql.DB) domain.OrderRepository {
	return &orderRepository{
		db: goqu.New("default", con),
	}
}
func (o *orderRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return withTransaction(ctx, o.db, fn)
}

func (o *orderRepository) FindAllByUserId(ctx context.Context, userId string) ([]domain.Order, error) {
	dataset := getExecutor(ctx, o.db).From("orders").Where(goqu.C("user_id").Eq(userId)).Order(goqu.C("created_at").Asc())
	var result []domain.Order
	err := dataset.ScanStructsContext(ctx, &result)
	if err != nil {
		return []domain.Order{}, err
	}
	return result, nil
}

// FindByIdAndUserId implements [domain.OrderRepository].
func (o *orderRepository) FindByIdAndUserId(ctx context.Context, id string, userId string) (domain.Order, error) {
	dataset := getExecutor(ctx, o.db).From("orders").Where(goqu.C("id").Eq(id), goqu.C("user_id").Eq(userId))
	var result domain.Order
	_, err := dataset.ScanStructContext(ctx, &result)
	if err != nil {
		return domain.Order{}, err
	}
	return result, nil
}

// FindById implements [domain.OrderRepository].
func (o *orderRepository) FindById(ctx context.Context, id string) (domain.Order, error) {
	panic("unimplemented")
}
func (o *orderRepository) FindByAll(ctx context.Context) ([]domain.Order, error) {
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
