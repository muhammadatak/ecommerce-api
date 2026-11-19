package service

import (
	"context"
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
	txRepository        domain.TxRepository
}

func NewOrder(
	orderRepository domain.OrderRepository,
	orderItemRepository domain.OrderItemRepository,
	cartRepository domain.CartRepository,
	cartItemRepository domain.CartItemRepository,
	productRepository domain.ProductRepository,
	txRepository domain.TxRepository,

) domain.OrderService {
	return &orderService{
		orderRepository:     orderRepository,
		orderItemRepository: orderItemRepository,
		cartRepository:      cartRepository,
		cartItemRepository:  cartItemRepository,
		productRepository:   productRepository,
		txRepository:        txRepository,
	}
}

func (o *orderService) Checkout(ctx context.Context, req dto.CheckoutRequest) error {
	return o.txRepository.WithTransaction(ctx, func(txCtx context.Context) error {
		cart, err := o.cartRepository.FindByUserId(txCtx, req.UserId)
		if err != nil {
			return err
		}

		if cart.Id == "" {
			return domain.ErrNotFound
		}

		items, err := o.cartItemRepository.FindAllByCartId(txCtx, cart.Id)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return domain.ErrEmptyCartItem
		}

		itemByProduct := make(map[string]domain.CartItem, 0)
		productIds := make([]string, 0)

		for _, v := range items {
			itemByProduct[v.ProductId] = v
			productIds = append(productIds, v.ProductId)
		}

		products, err := o.productRepository.FindByIds(txCtx, productIds)

		if err != nil {
			return err
		}
		if len(products) != len(items) {
			return domain.ErrNotFound
		}
		newOrderId := uuid.NewString()
		orderItems := make([]domain.OrderItem, 0)
		totalPrice := 0
		// save to order item
		for _, v := range products {
			qty := itemByProduct[v.Id].Quantity

			if qty > v.Stock {
				return domain.ErrInsufficientStock
			}

			if v.Status == domain.ProductStockInactive {
				return domain.ErrProductInactive
			}

			orderItems = append(orderItems, domain.OrderItem{
				Id:          uuid.NewString(),
				OrderId:     newOrderId,
				ProductId:   v.Id,
				ProductName: v.Name,
				Price:       v.Price,
				Quantity:    qty,
				Subtotal:    qty * v.Price,
				CreatedAt:   time.Now(),
			})
			totalPrice += v.Price * qty
		}

		err = o.orderRepository.Save(txCtx, domain.Order{
			Id:         newOrderId,
			UserId:     req.UserId,
			Status:     domain.OrderStatusPending,
			TotalPrice: totalPrice,
			CreatedAt:  time.Now(),
		})
		if err != nil {
			return err
		}

		// save item ke order
		err = o.orderItemRepository.SaveBatch(txCtx, orderItems)
		if err != nil {
			return err
		}

		// minus stok produk by item quantity
		for _, v := range products {
			v.Stock -= itemByProduct[v.Id].Quantity
			err = o.productRepository.Update(txCtx, &v)
			if err != nil {
				return err
			}
		}

		err = o.cartItemRepository.DeleteByCartId(txCtx, cart.Id)
		if err != nil {
			return err
		}
		return nil
	})
}
