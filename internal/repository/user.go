package repository

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/muhammadatak/go-ecommerce-api/domain"
)

type userRepository struct {
	db *goqu.Database
}

func NewUser(con *sql.DB) domain.UserRepository {
	return &userRepository{
		db: goqu.New("default", con),
	}
}

// FindByEmail implements [domain.UserRepository].
func (u *userRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	dataset := u.db.From("users").Where(goqu.C("email").Eq(email))
	var result domain.User
	_, err := dataset.ScanStructContext(ctx, &result)

	if err != nil {
		return domain.User{}, err
	}
	return result, nil
}
