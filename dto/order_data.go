package dto

type CheckoutRequest struct {
	UserId string `json:"-"`
}

type CheckoutResponses struct {
	Id         string `json:"id"`
	Status     string `json:"status"`
	TotalPrice int    `json:"total_price"`
}
