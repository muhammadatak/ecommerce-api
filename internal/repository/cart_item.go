package repository

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/muhammadatak/go-ecommerce-api/domain"
)

type cartItemRepository struct {
	db *goqu.Database
}

func NewCartItem(con *sql.DB) domain.CartItemRepository {
	return &cartItemRepository{
		db: goqu.New("default", con),
	}
}

// Create implements [domain.CartItemRepository].
func (c *cartItemRepository) Save(ctx context.Context, item domain.CartItem) error {
	executor := c.db.Insert("cart_items").Rows(item).Executor()
	_, err := executor.ExecContext(ctx)
	return err
}

func (c *cartItemRepository) FindByCartId(ctx context.Context, cartId string) ([]domain.CartItem, error) {
	dataset := c.db.From("cart_items").Where(goqu.C("cart_id").Eq(cartId))
	var result []domain.CartItem
	err := dataset.ScanStructsContext(ctx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
