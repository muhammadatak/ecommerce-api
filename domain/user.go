package domain

import (
	"context"

	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type User struct {
	Id       string `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (User, error)
}

type UserService interface {
	Get(ctx context.Context, email string) (dto.UserData, error)
}
