package domain

import "errors"

var (
	ErrNotFound          = errors.New("Data is not found")
	ErrProductInactive   = errors.New("Product is inactive")
	ErrInsufficientStock = errors.New("Insufficient stock")
	ErrAuth              = errors.New("Wrong email or password")
	ErrInvalidToken      = errors.New("Invalid token")
)
