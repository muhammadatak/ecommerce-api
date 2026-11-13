package repository

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/muhammadatak/go-ecommerce-api/domain"
)

type cartRepository struct {
	db *goqu.Database
}

func NewCart(con *sql.DB) domain.CartRepository {
	return &cartRepository{
		db: goqu.New("default", con),
	}
}

// FindById implements [domain.CartRepository].
func (c *cartRepository) FindByUserId(ctx context.Context, userId string) (domain.Cart, error) {
	dataset := c.db.From("carts").Where(goqu.C("user_id").Eq(userId))
	var result domain.Cart
	_, err := dataset.ScanStructContext(ctx, &result)
	if err != nil {
		return domain.Cart{}, err
	}
	return result, nil
}

// Create implements [domain.CartRepository].
func (c *cartRepository) Save(ctx context.Context, cart domain.Cart) error {
	executor := c.db.Insert("carts").Rows(cart).Executor()
	_, err := executor.ExecContext(ctx)

	return err
}
