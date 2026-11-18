package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type orderService struct {
	orderRepository     domain.OrderRepository
	orderItemRepository domain.OrderItemRepository
	cartRepository      domain.CartRepository
	cartItemRepository  domain.CartItemRepository
	productRepository   domain.ProductRepository
}

func NewOrder(
	orderRepository domain.OrderRepository,
	orderItemRepository domain.OrderItemRepository,
	cartRepository domain.CartRepository,
	cartItemRepository domain.CartItemRepository,
	productRepository domain.ProductRepository,

) domain.OrderService {
	return &orderService{
		orderRepository:     orderRepository,
		orderItemRepository: orderItemRepository,
		cartRepository:      cartRepository,
		cartItemRepository:  cartItemRepository,
		productRepository:   productRepository,
	}
}

func (o *orderService) Checkout(ctx context.Context, req dto.CheckoutRequest) error {
	err := domain.RunInTransaction(ctx, o.orderRepository, func(txCtx context.Context) error {
		cart, err := o.cartRepository.FindByUserId(txCtx, req.UserId)
		if err != nil {
			return err
		}

		if cart.Id == "" {
			return domain.ErrNotFound
		}
		fmt.Printf("product = %s\n", cart.Id)

		items, err := o.cartItemRepository.FindAllByCartId(txCtx, cart.Id)

		if err != nil {
			return err
		}
		if len(items) == 0 {
			return errors.New("cart item empty")
		}
		fmt.Printf("item len = %v\n", len(items))

		productIds := make([]string, 0)
		for _, v := range items {
			productIds = append(productIds, v.ProductId)
		}
		products, err := o.productRepository.FindByIds(txCtx, productIds)
		for _, v := range products {
			fmt.Printf("product = %s\n", v.Name)
		}

		if err != nil {
			return err
		}

		// create order
		newOrder := domain.Order{
			Id:         uuid.NewString(),
			UserId:     req.UserId,
			Status:     domain.OrderStatusPending,
			TotalPrice: 0,
			CreatedAt:  time.Now(),
		}
		err = o.orderRepository.Save(txCtx, newOrder)
		if err != nil {
			return err
		}
		totalPrice := 0
		// save to order item
		for _, v := range products {
			qtyItem, err := o.cartItemRepository.FindByCartAndProduct(txCtx, cart.Id, v.Id)

			if err != nil {
				return err
			}
			qty := qtyItem.Quantity

			if qty > v.Stock {
				return domain.ErrInsufficientStock
			}

			if v.Status == domain.ProductStockInactive {
				return domain.ErrProductInactive
			}

			err = o.orderItemRepository.Save(txCtx, &domain.OrderItem{
				Id:          uuid.NewString(),
				OrderId:     newOrder.Id,
				ProductId:   v.Id,
				ProductName: v.Name,
				Price:       v.Price,
				Quantity:    qty,
				Subtotal:    qty * v.Price,
				CreatedAt:   time.Now(),
			})

			if err != nil {
				return err
			}
			v.Stock = v.Stock - qty
			err = o.productRepository.Update(txCtx, &v)
			if err != nil {
				return err
			}
			fmt.Printf("%s stock = %v\n", v.Name, v.Stock)

			totalPrice += v.Price * qty
			fmt.Println("ok")
			err = o.cartItemRepository.Delete(txCtx, qtyItem.Id)
			if err != nil {
				return err
			}
		}

		newOrder = domain.Order{
			Id:         newOrder.Id,
			TotalPrice: totalPrice,
		}
		err = o.orderRepository.Update(txCtx, &newOrder)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
