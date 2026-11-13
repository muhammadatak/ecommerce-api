package api

import (
	"context"
	"net/http"
	"time"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
	"github.com/muhammadatak/go-ecommerce-api/internal/util"
)

type cartApi struct {
	cartService domain.CartService
}

func NewCart(app *fiber.App, cartService domain.CartService, authMid fiber.Handler) {
	ca := cartApi{
		cartService: cartService,
	}

	carts := app.Group("/carts", authMid)

	carts.Get("/", ca.Show)
	carts.Post("/item", ca.AddItem)

}

func (ca cartApi) Show(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	token := jwtware.FromContext(ctx)
	if token == nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.CreateResponseError("token tidak valid"))
	}

	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.CreateResponseError("token tidak valid"))
	}

	userId, ok := claim["id"].(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.CreateResponseError("token tidak valid"))
	}

	res, err := ca.cartService.Show(c, userId)

	if err != nil {
		return ctx.Status(http.StatusInternalServerError).
			JSON(dto.CreateResponseError(err.Error()))
	}

	return ctx.JSON(dto.CreateResponseSuccess(res))

}

func (ca cartApi) AddItem(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	var req dto.CreateItemRequest
	err := ctx.Bind().Body(&req)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).
			JSON(dto.CreateResponseError(err.Error()))
	}

	fails := util.Validate(req)
	if len(fails) > 0 {
		return ctx.Status(http.StatusBadRequest).JSON(dto.CreateResponseErrorData("Validation failed", fails))
	}

	token := jwtware.FromContext(ctx)
	if token == nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.CreateResponseError("token tidak valid"))
	}

	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.CreateResponseError("token tidak valid"))
	}

	userId, ok := claim["id"].(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.CreateResponseError("token tidak valid"))
	}
	req.UserId = userId
	err = ca.cartService.AddItem(c, req)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).
			JSON(dto.CreateResponseError(err.Error()))
	}
	return ctx.SendStatus(http.StatusCreated)
}
