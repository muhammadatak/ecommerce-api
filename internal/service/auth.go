package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
	"github.com/muhammadatak/go-ecommerce-api/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	cnf            *config.Config
	userRepository domain.UserRepository
}

func NewAuth(userRepository domain.UserRepository, cnf *config.Config) domain.AuthService {
	return &authService{
		userRepository: userRepository,
		cnf:            cnf,
	}
}

// Login implements [domain.AuthService].
func (a *authService) Login(ctx context.Context, req dto.AuthRequest) (dto.AuthResponse, error) {
	user, err := a.userRepository.FindByEmail(ctx, req.Email)

	if err != nil {
		return dto.AuthResponse{}, err
	}
	if user.Id == "" {
		return dto.AuthResponse{}, errors.New("email atau password salah")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return dto.AuthResponse{}, errors.New("email atau password salah")
	}

	claim := jwt.MapClaims{
		"id":  user.Id,
		"exp": time.Now().Add(time.Duration(a.cnf.Jwt.Exp) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenStr, err := token.SignedString([]byte(a.cnf.Jwt.Key))

	if err != nil {
		return dto.AuthResponse{}, errors.New("auth gagal")
	}

	return dto.AuthResponse{
		Token: tokenStr,
	}, nil
}
