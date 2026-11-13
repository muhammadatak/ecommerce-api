package api

import (
	"errors"
	"net/http"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/muhammadatak/go-ecommerce-api/domain"
)

func errorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrAuth):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrInsufficientStock),
		errors.Is(err, domain.ErrProductInactive),
		errors.Is(err, domain.ErrInvalidToken):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func getUserId(ctx fiber.Ctx) (string, error) {
	token := jwtware.FromContext(ctx)
	if token == nil {
		return "", domain.ErrInvalidToken
	}

	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", domain.ErrInvalidToken
	}

	userId, ok := claim["id"].(string)
	if !ok {
		return "", domain.ErrInvalidToken
	}
	return userId, nil
}
