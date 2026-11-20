package dto

import "time"

type CheckoutRequest struct {
	UserId string `json:"-"`
}

type OrderData struct {
	Id         string    `json:"id"`
	Status     string    `json:"status"`
	TotalPrice int       `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
}

type OrderItem struct {
	ProductName string `json:"product_name"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
	Subtotal    int    `json:"subtotal"`
}

type OrderDetail struct {
	Order OrderData   `json:"order"`
	Items []OrderItem `json:"items"`
}
