package domain

import (
	"context"
	"database/sql"

	"github.com/muhammadatak/go-ecommerce-api/dto"
)

const (
	ProductStockActive   = "ACTIVE"
	ProductStockInactive = "INACTIVE"
)

type Product struct {
	Id          string       `db:"id"`
	Name        string       `db:"name"`
	Description string       `db:"description"`
	Price       int          `db:"price"`
	Stock       int          `db:"stock"`
	Status      string       `db:"status"`
	CreatedAt   sql.NullTime `db:"created_at"`
}

type ProductRepository interface {
	FindAll(ctx context.Context) ([]Product, error)
	FindById(ctx context.Context, id string) (Product, error)
	FindByIds(ctx context.Context, ids []string) ([]Product, error)
	Save(ctx context.Context, product *Product) error
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id string) error
}

type ProductService interface {
	Index(ctx context.Context) ([]dto.ProductData, error)
	Show(ctx context.Context, id string) (dto.ProductData, error)
	Create(ctx context.Context, req dto.CreateProductReq) (dto.ProductData, error)
	Update(ctx context.Context, req dto.UpdateProductReq) (dto.ProductData, error)
	Delete(ctx context.Context, id string) error
}
