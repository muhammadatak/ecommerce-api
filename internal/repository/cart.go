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

// Delete implements [domain.CartRepository].

// FindAll implements [domain.CartRepository].
func (c *cartRepository) FindAll(ctx context.Context) ([]domain.Cart, error) {
	panic("unimplemented")
}

// FindById implements [domain.CartRepository].
func (c *cartRepository) FindById(ctx context.Context, id string) (domain.Cart, error) {
	dataset := c.db.From("carts").Where(goqu.C("id").Eq(id))
	var result domain.Cart
	_, err := dataset.ScanStructContext(ctx, &result)
	if err != nil {
		return domain.Cart{}, err
	}
	return result, nil
}

// Save implements [domain.CartRepository].
func (c *cartRepository) Save(ctx context.Context, cart domain.Cart) error {
	panic("unimplemented")
}

func (c *cartRepository) Delete(Ctx context.Context, id string) error {
	panic("unimplemented")
}
