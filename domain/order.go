package domain

import (
	"context"
	"database/sql"
	"time"

	"github.com/muhammadatak/go-ecommerce-api/dto"
)

const (
	OrderStatusPending   = "PENDING"
	OrderStatusPaid      = "PAID"
	OrderStatusShipped   = "SHIPPED"
	OrderStatusCompleted = "COMPLETED"
	OrderStatusCancelled = "CANCELLED"
)

type Order struct {
	Id         string       `db:"id"`
	UserId     string       `db:"user_id"`
	Status     string       `db:"status"`
	TotalPrice int          `db:"total_price"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  sql.NullTime `db:"updated_at"`
}

type OrderRepository interface {
	FindById(ctx context.Context, id string) (Order, error)
	FindByAll(ctx context.Context) ([]Order, error)
	Save(ctx context.Context, order Order) error
	Update(ctx context.Context, order *Order) error
	FindAllByUserId(ctx context.Context, userId string) ([]Order, error)
	FindByIdAndUserId(ctx context.Context, id string, userId string) (Order, error)
}

type OrderService interface {
	Checkout(ctx context.Context, req dto.CheckoutRequest) (dto.OrderData, error)
	Index(ctx context.Context, userId string) ([]dto.OrderData, error)
	Show(ctx context.Context, id string, userId string) (dto.OrderDetail, error)
}
