package service

import (
	"context"

	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type cartService struct {
	cartRepository     domain.CartRepository
	productRepository  domain.ProductRepository
	cartItemRepository domain.CartItemRepository
}

func NewCart(
	cartRepository domain.CartRepository,
	productRepository domain.ProductRepository,
	cartItemRepository domain.CartItemRepository) domain.CartService {
	return &cartService{
		cartRepository:     cartRepository,
		productRepository:  productRepository,
		cartItemRepository: cartItemRepository,
	}
}

// Show implements [domain.CartService].
func (c *cartService) Show(ctx context.Context, userId string) (dto.CartData, error) {
	cart, err := c.cartRepository.FindByUserId(ctx, userId)

	if err != nil {
		return dto.CartData{}, err
	}
	if cart.Id == "" {
		return dto.CartData{CartItem: []dto.CartItemData{}}, err
	}
	cartItems, err := c.cartItemRepository.FindByCartId(ctx, cart.Id)

	if err != nil {
		return dto.CartData{}, err
	}

	cartItemsData := make([]dto.CartItemData, 0)
	for _, v := range cartItems {
		productDb, err := c.productRepository.FindById(ctx, v.ProductId)
		if err != nil {
			return dto.CartData{}, err
		}
		cartItemsData = append(cartItemsData, dto.CartItemData{
			Id: v.Id,
			Product: dto.ProductData{
				Id:          v.ProductId,
				Name:        productDb.Name,
				Description: productDb.Description,
				Price:       productDb.Price,
				Stock:       productDb.Stock,
				Status:      productDb.Status,
			},
			Quantity: v.Quantity,
		})
	}
	return dto.CartData{
		Id:       cart.Id,
		CartItem: cartItemsData,
	}, nil

}
