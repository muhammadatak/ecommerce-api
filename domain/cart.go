package domain

import (
	"context"
	"database/sql"

	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type Cart struct {
	Id        string       `db:"id"`
	UserId    string       `db:"user_id"`
	CreatedAt sql.NullTime `db:"created_at"`
}

type CartRepository interface {
	FindByUserId(ctx context.Context, UserId string) (Cart, error)
	Save(ctx context.Context, cart Cart) error
}

type CartService interface {
	Show(ctx context.Context, UserId string) (dto.CartData, error)
	AddItem(ctx context.Context, req dto.CreateItemRequest) error
}
