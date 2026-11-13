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

type cartItemService struct {
	cartItemRepository domain.CartItemRepository
	cartRepository     domain.CartRepository
	productRepository  domain.ProductRepository
}

func NewCartItem(
	cartItemRepository domain.CartItemRepository,
	cartRepository domain.CartRepository,
	productRepository domain.ProductRepository) domain.CartItemService {
	return &cartItemService{
		cartItemRepository: cartItemRepository,
		cartRepository:     cartRepository,
		productRepository:  productRepository,
	}
}

func (c *cartItemService) Create(ctx context.Context, req dto.CreateItemRequest) error {

	product, err := c.productRepository.FindById(ctx, req.ProductId)
	if err != nil {
		return err
	}
	if product.Id == "" {
		return errors.New("product not found")
	}

	cart, err := c.cartRepository.FindByUserId(ctx, req.UserId)
	if err != nil {
		return err
	}
	if cart.Id == "" {
		return errors.New("product not found")
	}
	item := domain.CartItem{
		Id:        uuid.NewString(),
		CartId:    cart.Id,
		ProductId: req.ProductId,
		Quantity:  req.Quantity,
		CreatedAt: sql.NullTime{Valid: true, Time: time.Now()},
	}
	err = c.cartItemRepository.Save(ctx, item)
	return err
}
