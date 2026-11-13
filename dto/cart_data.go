package dto

type CartItemData struct {
	Id       string      `json:"id"`
	Product  ProductData `json:"product"`
	Quantity int         `json:"quantity"`
}

type CartData struct {
	Id       string         `json:"id"`
	CartItem []CartItemData `json:"cart_item"`
}
