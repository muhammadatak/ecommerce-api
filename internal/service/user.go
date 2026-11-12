package service

import (
	"context"
	"errors"

	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type userService struct {
	userRepository domain.UserRepository
}

func NewUser(userRepository domain.UserRepository) domain.UserService {
	return &userService{
		userRepository: userRepository,
	}
}

// Get implements [domain.UserService].
func (u *userService) Get(ctx context.Context, email string) (dto.UserData, error) {
	user, err := u.userRepository.FindByEmail(ctx, email)

	if err != nil {
		return dto.UserData{}, err
	}

	if user.Id == "" {
		return dto.UserData{}, errors.New("User not found")
	}

	return dto.UserData{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}
