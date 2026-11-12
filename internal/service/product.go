package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type productService struct {
	productRepository domain.ProductRepository
}

func NewProduct(productRepository domain.ProductRepository) domain.ProductService {
	return &productService{
		productRepository: productRepository,
	}
}

// Create implements [domain.ProductService].
func (p *productService) Create(ctx context.Context, req dto.CreateProductReq) (dto.ProductData, error) {
	data := domain.Product{
		Id:          uuid.NewString(),
		Name:        req.Name,
		Description: req.Description,
		Stock:       req.Stock,
		Price:       req.Price,
		Status:      domain.ProductStockActive,
		CreatedAt:   sql.NullTime{Valid: true, Time: time.Now()},
	}
	err := p.productRepository.Save(ctx, &data)
	if err != nil {
		return dto.ProductData{}, err
	}
	return dto.ProductData{
		Id:          data.Id,
		Name:        data.Name,
		Description: data.Description,
		Price:       data.Price,
		Stock:       data.Stock,
		Status:      data.Status,
	}, nil
}

// Index implements [domain.ProductService].
func (p *productService) Index(ctx context.Context) ([]dto.ProductData, error) {
	products, err := p.productRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var productData []dto.ProductData

	for _, v := range products {
		productData = append(productData, dto.ProductData{
			Id:          v.Id,
			Name:        v.Name,
			Description: v.Description,
			Price:       v.Price,
			Stock:       v.Stock,
			Status:      v.Status,
		})
	}
	return productData, nil
}

// Show implements [domain.ProductService].
func (p *productService) Show(ctx context.Context, id string) (dto.ProductData, error) {
	product, err := p.productRepository.FindById(ctx, id)
	if err != nil {
		return dto.ProductData{}, err
	}
	if product.Id == "" {
		return dto.ProductData{}, errors.New("Product not found")
	}

	return dto.ProductData{
		Id:          product.Id,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		Status:      product.Status,
	}, nil

}

// Update implements [domain.ProductService].
func (p *productService) Update(ctx context.Context, req dto.UpdateProductReq) (dto.ProductData, error) {
	exists, err := p.productRepository.FindById(ctx, req.Id)
	if err != nil {
		return dto.ProductData{}, err
	}
	if exists.Id == "" {
		return dto.ProductData{}, errors.New("Product not found")

	}

	if req.Name != nil {
		exists.Name = *req.Name
	}
	if req.Description != nil {
		exists.Description = *req.Description
	}
	if req.Stock != nil {
		exists.Stock = *req.Stock
	}
	if req.Price != nil {
		exists.Price = *req.Price
	}
	if req.Status != nil {
		exists.Status = *req.Status
	}

	err = p.productRepository.Update(ctx, &exists)
	if err != nil {
		return dto.ProductData{}, err
	}

	return dto.ProductData{
		Id:          exists.Id,
		Name:        exists.Name,
		Description: exists.Description,
		Price:       exists.Price,
		Stock:       exists.Stock,
		Status:      exists.Status,
	}, nil
}

// Delete implements [domain.ProductService].
func (p *productService) Delete(ctx context.Context, id string) error {
	exists, err := p.productRepository.FindById(ctx, id)
	if err != nil {
		return err
	}
	if exists.Id == "" {
		return errors.New("Product not found")
	}

	err = p.productRepository.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil

}
