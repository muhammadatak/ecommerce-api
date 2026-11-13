package dto

type CreateItemRequest struct {
	UserId    string `json:"-"`
	ProductId string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}
