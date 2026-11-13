package dto

type CreateItemRequest struct {
	UserId    string `json:"-"`
	ProductId string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
