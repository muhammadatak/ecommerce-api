package repository

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/muhammadatak/go-ecommerce-api/domain"
)

type orderItemRepository struct {
	db *goqu.Database
}

func NewOrderItem(con *sql.DB) domain.OrderItemRepository {
	return &orderItemRepository{
		db: goqu.New("default", con),
	}
}

// FindAllByOrderId implements [domain.OrderItemRepository].
func (o *orderItemRepository) FindAllByOrderId(ctx context.Context, orderId string) ([]domain.OrderItem, error) {
	dataset := getExecutor(ctx, o.db).From("order_items").Where(goqu.C("order_id").Eq(orderId))
	var result []domain.OrderItem
	err := dataset.ScanStructsContext(ctx, &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Save implements [domain.OrderItemRepository].
func (o *orderItemRepository) Save(ctx context.Context, item *domain.OrderItem) error {
	executor := getExecutor(ctx, o.db).Insert("order_items").Rows(item).Executor()
	_, err := executor.ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (o *orderItemRepository) Delete(ctx context.Context, id string) error {
	executor := getExecutor(ctx, o.db).Delete("order_items").Where(goqu.C("id").Eq(id)).Executor()
	_, err := executor.ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil

}
