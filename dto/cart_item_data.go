package dto

type CreateItemRequest struct {
	UserId    string `json:"-"`
	ProductId string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}
type UpdateItemQuantity struct {
	Id       string `json:"-"`
	UserId   string `json:"-"`
	Quantity int    `json:"quantity" validate:"gt=0"`
}
