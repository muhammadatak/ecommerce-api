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

// FindByCartId implements [domain.CartItemRepository].

func NewCartItem(con *sql.DB) domain.CartItemRepository {
	return &cartItemRepository{
		db: goqu.New("default", con),
	}
}
func (c *cartItemRepository) FindByIdAndCartId(ctx context.Context, id string, cartId string) (domain.CartItem, error) {
	dataset := getExecutor(ctx, c.db).From("cart_items").Where(goqu.C("id").Eq(id), goqu.C("cart_id").Eq(cartId))
	var result domain.CartItem
	_, err := dataset.ScanStructContext(ctx, &result)
	if err != nil {
		return domain.CartItem{}, err
	}

	return result, nil
}

// Create implements [domain.CartItemRepository].
func (c *cartItemRepository) Save(ctx context.Context, item domain.CartItem) error {
	executor := getExecutor(ctx, c.db).Insert("cart_items").Rows(item).
		OnConflict(goqu.DoUpdate("cart_id, product_id",
			goqu.Record{"quantity": goqu.L("cart_items.quantity + EXCLUDED.quantity"),
				"updated_at": goqu.L("NOW()")})).Executor()
	_, err := executor.ExecContext(ctx)
	return err
}

func (c *cartItemRepository) FindAllByCartId(ctx context.Context, cartId string) ([]domain.CartItem, error) {
	dataset := getExecutor(ctx, c.db).From("cart_items").Where(goqu.C("cart_id").Eq(cartId))
	var result []domain.CartItem
	err := dataset.ScanStructsContext(ctx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *cartItemRepository) FindByCartAndProduct(ctx context.Context, cartId string, productId string) (domain.CartItem, error) {
	dataset := getExecutor(ctx, c.db).From("cart_items").Where(goqu.C("cart_id").Eq(cartId), goqu.C("product_id").Eq(productId))
	var result domain.CartItem
	_, err := dataset.ScanStructContext(ctx, &result)
	if err != nil {
		return domain.CartItem{}, err
	}

	return result, nil
}

// Delete implements [domain.CartItemRepository].
func (c *cartItemRepository) Delete(ctx context.Context, id string) error {
	executor := getExecutor(ctx, c.db).Delete("cart_items").Where(goqu.C("id").Eq(id)).Executor()
	_, err := executor.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Update implements [domain.CartItemRepository].
func (c *cartItemRepository) Update(ctx context.Context, item *domain.CartItem) error {
	executor := getExecutor(ctx, c.db).Update("cart_items").Where(goqu.C("id").Eq(item.Id), goqu.C("cart_id").Eq(item.CartId)).
		Set(goqu.Record{"quantity": item.Quantity, "updated_at": item.UpdatedAt}).Executor()
	_, err := executor.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *cartItemRepository) DeleteByCartId(ctx context.Context, cartId string) error {
	executor := getExecutor(ctx, c.db).Delete("cart_items").Where(goqu.C("cart_id").Eq(cartId)).Executor()
	_, err := executor.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}
