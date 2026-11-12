package dto

type ProductData struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`
	Status      string `json:"status"`
}

type CreateProductReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`
}

type UpdateProductReq struct {
	Id          string  `json:"-"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int    `json:"price"`
	Stock       *int    `json:"stock"`
	Status      *string `json:"status"`
}
