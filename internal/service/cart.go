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

	productsIds := make([]string, 0)
	for _, v := range cartItems {
		productsIds = append(productsIds, v.ProductId)
	}
	products, err := c.productRepository.FindByIds(ctx, productsIds)
	if err != nil {
		return dto.CartData{}, err
	}

	productMap := make(map[string]domain.Product, 0)
	for _, v := range products {
		productMap[v.Id] = v
	}

	cartItemsData := make([]dto.CartItemData, 0)
	for _, v := range cartItems {
		product := productMap[v.ProductId]
		cartItemsData = append(cartItemsData, dto.CartItemData{
			Id: v.Id,
			Product: dto.ProductData{
				Id:          v.ProductId,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Stock:       product.Stock,
				Status:      product.Status,
			},
			Quantity: v.Quantity,
		})

	}
	return dto.CartData{
		Id:       cart.Id,
		CartItem: cartItemsData,
	}, nil
}

func (c *cartService) Create(ctx context.Context, req dto.CreateItemRequest) error {

	product, err := c.productRepository.FindById(ctx, req.ProductId)
	if err != nil {
		return err
	}
	if product.Id == "" {
		return errors.New("product not found")
	}
	if product.Status == domain.ProductStockInactive {
		return errors.New("Product status is not active")
	}
	if req.Quantity > product.Stock {
		return errors.New("Product stock not cukupz ")
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
