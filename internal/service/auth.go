package service

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := a.userRepository.FindByEmail(ctx, normalizedEmail)

	if err != nil {
		return dto.AuthResponse{}, err
	}
	if user.Id == "" {
		return dto.AuthResponse{}, domain.ErrAuth
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return dto.AuthResponse{}, domain.ErrAuth
	}

	token, err := issueToken(user.Id, a.cnf)

	if err != nil {
		return dto.AuthResponse{}, err
	}

	return dto.AuthResponse{
		Token: token,
	}, nil
}

func (a *authService) Register(ctx context.Context, req dto.RegisterRequest) (dto.AuthResponse, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := a.userRepository.FindByEmail(ctx, normalizedEmail)

	if err != nil {
		return dto.AuthResponse{}, err
	}

	if strings.EqualFold(user.Email, normalizedEmail) {
		return dto.AuthResponse{}, domain.ErrDupeEmail
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		return dto.AuthResponse{}, err
	}

	newUserId := uuid.NewString()
	err = a.userRepository.Save(ctx, domain.User{
		Id:       newUserId,
		Email:    normalizedEmail,
		Name:     req.Name,
		Password: string(hashPassword),
	})

	if err != nil {
		return dto.AuthResponse{}, err
	}

	token, err := issueToken(newUserId, a.cnf)

	if err != nil {
		return dto.AuthResponse{}, err
	}

	return dto.AuthResponse{
		Token: token,
	}, nil
}

func issueToken(userId string, cnf *config.Config) (string, error) {
	claim := jwt.MapClaims{
		"id":  userId,
		"exp": time.Now().Add(time.Duration(cnf.Jwt.Exp) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenStr, err := token.SignedString([]byte(cnf.Jwt.Key))

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
