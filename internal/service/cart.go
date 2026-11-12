package service

import (
	"context"

	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type cartService struct {
	cartRepository    domain.CartRepository
	productRepository domain.ProductRepository
}

func NewCart(cartRepository domain.CartRepository) domain.CartService {
	return &cartService{
		cartRepository: cartRepository,
	}
}

func (c *cartService) Show(ctx context.Context) (dto.CartData, error) {
	panic("none")
}
