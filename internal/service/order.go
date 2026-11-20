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

// Index implements [domain.OrderService].
func (o *orderService) Index(ctx context.Context, userId string) ([]dto.OrderData, error) {
	orders, err := o.orderRepository.FindAllByUserId(ctx, userId)
	if err != nil {
		return []dto.OrderData{}, err
	}

	orderData := make([]dto.OrderData, 0)
	for _, v := range orders {
		orderData = append(orderData, dto.OrderData{
			Id:         v.Id,
			Status:     v.Status,
			TotalPrice: v.TotalPrice,
			CreatedAt:  v.CreatedAt,
		})
	}
	return orderData, nil
}

// Show implements [domain.OrderService].
func (o *orderService) Show(ctx context.Context, id string, userId string) (dto.OrderDetail, error) {
	order, err := o.orderRepository.FindByIdAndUserId(ctx, id, userId)
	if err != nil {
		return dto.OrderDetail{}, err
	}
	if order.Id == "" {
		return dto.OrderDetail{}, domain.ErrNotFound
	}

	items, err := o.orderItemRepository.FindAllByOrderId(ctx, order.Id)
	if err != nil {
		return dto.OrderDetail{}, err
	}
	itemData := make([]dto.OrderItem, 0)
	for _, v := range items {
		itemData = append(itemData, dto.OrderItem{
			ProductName: v.ProductName,
			Price:       v.Price,
			Quantity:    v.Quantity,
			Subtotal:    v.Subtotal,
		})
	}

	return dto.OrderDetail{
		Order: dto.OrderData{
			Id:         order.Id,
			Status:     order.Status,
			TotalPrice: order.TotalPrice,
			CreatedAt:  order.CreatedAt,
		},
		Items: itemData,
	}, nil
}

func (o *orderService) Checkout(ctx context.Context, req dto.CheckoutRequest) (dto.OrderData, error) {
	var order domain.Order
	err := o.txRepository.WithTransaction(ctx, func(txCtx context.Context) error {
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

		order = domain.Order{
			Id:         newOrderId,
			UserId:     req.UserId,
			Status:     domain.OrderStatusPending,
			TotalPrice: totalPrice,
			CreatedAt:  time.Now(),
		}
		err = o.orderRepository.Save(txCtx, order)
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
			err = o.productRepository.UpdateStock(txCtx, &v)
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
	if err != nil {
		return dto.OrderData{}, err
	}
	res := dto.OrderData{
		Id:         order.Id,
		Status:     order.Status,
		TotalPrice: order.TotalPrice,
		CreatedAt:  order.CreatedAt,
	}
	return res, nil
}
