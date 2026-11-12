package repository

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/muhammadatak/go-ecommerce-api/domain"
)

type productRepository struct {
	db *goqu.Database
}

func NewProduct(con *sql.DB) domain.ProductRepository {
	return &productRepository{
		db: goqu.New("default", con),
	}
}

// FindAll implements [domain.ProductRepository].
func (p *productRepository) FindAll(ctx context.Context) ([]domain.Product, error) {
	dataset := p.db.From("products")
	var result []domain.Product
	err := dataset.ScanStructsContext(ctx, &result)

	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindById implements [domain.ProductRepository].
func (p *productRepository) FindById(ctx context.Context, id string) (domain.Product, error) {
	dataset := p.db.From("products").Where(goqu.C("id").Eq(id))
	var result domain.Product

	_, err := dataset.ScanStructContext(ctx, &result)
	if err != nil {
		return domain.Product{}, err
	}
	return result, nil
}

// Save implements [domain.ProductRepository].
func (p *productRepository) Save(ctx context.Context, product *domain.Product) error {
	executor := p.db.Insert("products").Rows(product).Executor()
	_, err := executor.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Update implements [domain.ProductRepository].
func (p *productRepository) Update(ctx context.Context, product *domain.Product) error {
	executor := p.db.Update("products").Where(goqu.C("id").Eq(product.Id)).Set(product).Executor()
	_, err := executor.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Delete implements [domain.ProductRepository].
func (p *productRepository) Delete(ctx context.Context, id string) error {
	executor := p.db.Delete("products").Where(goqu.C("id").Eq(id)).Executor()
	_, err := executor.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}
