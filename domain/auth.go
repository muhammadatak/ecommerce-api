package domain

import (
	"context"

	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type AuthService interface {
	Login(ctx context.Context, req dto.AuthRequest) (dto.AuthResponse, error)
}
