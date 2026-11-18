package dto

type CheckoutRequest struct {
	UserId string `json:"-"`
	CartId string `json:"cart_id" validate:"required,uuid"`
}
