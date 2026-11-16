package domain

import (
	"context"

	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type User struct {
	Id       string `db:"id"`
	Email    string `db:"email"`
	Name     string `db:"name"`
	Password string `db:"password"`
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	Save(ctx context.Context, user User) error
}

type UserService interface {
	Get(ctx context.Context, email string) (dto.UserData, error)
}
