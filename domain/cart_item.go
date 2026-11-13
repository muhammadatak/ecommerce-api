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
	FindByCartId(ctx context.Context, cartId string) ([]CartItem, error)
	FindByCartAndProduct(ctx context.Context, cartId string, productId string) (CartItem, error)
	Save(ctx context.Context, item CartItem) error
}
