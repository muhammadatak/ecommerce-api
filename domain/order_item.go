package domain

import (
	"context"
	"time"
)

type OrderItem struct {
	Id          string    `db:"id"`
	OrderId     string    `db:"order_id"`
	ProductId   string    `db:"product_id"`
	ProductName string    `db:"product_name"`
	Price       int       `db:"price"`
	Quantity    int       `db:"quantity"`
	Subtotal    int       `db:"subtotal"`
	CreatedAt   time.Time `db:"created_at"`
}

type OrderItemRepository interface {
	FindAllByOrderId(ctx context.Context, id string) ([]OrderItem, error)
	Save(ctx context.Context, item *OrderItem) error
	Delete(ctx context.Context, id string) error
}
