package domain

import (
	"context"
	"database/sql"
)

type CartItem struct {
	Id        string       `db:"id"`
	CartId    string       `db:"cart_id"`
	ProductId string       `db:"product_id"`
	Quantity  int          `db:"quantity"`
	UpdatedAt sql.NullTime `db:"updated_at"`
	CreatedAt sql.NullTime `db:"created_at"`
}

type CartItemRepository interface {
	FindAll(ctx context.Context) ([]CartItem, error)
}
