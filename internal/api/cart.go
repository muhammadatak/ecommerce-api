package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
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
	carts.Patch("/item/:id", ca.UpdateItemQuantity)
	carts.Delete("/item/:id", ca.DeleteItem)

}

func (ca cartApi) Show(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	userId, err := getUserId(ctx)
	if err != nil {
		return ctx.Status(errorStatus(err)).JSON(dto.CreateResponseError(err.Error()))
	}

	res, err := ca.cartService.Show(c, userId)

	if err != nil {
		return ctx.Status(errorStatus(err)).
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
		return ctx.Status(http.StatusBadRequest).
			JSON(dto.CreateResponseError(err.Error()))
	}

	fails := util.Validate(req)
	if len(fails) > 0 {
		return ctx.Status(http.StatusBadRequest).JSON(dto.CreateResponseErrorData("Validation failed", fails))
	}

	userId, err := getUserId(ctx)
	if err != nil {
		return ctx.Status(errorStatus(err)).JSON(dto.CreateResponseError(err.Error()))
	}
	req.UserId = userId
	err = ca.cartService.AddItem(c, req)
	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}
	return ctx.SendStatus(http.StatusCreated)
}

func (ca cartApi) UpdateItemQuantity(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	var req dto.UpdateItemQuantity
	err := ctx.Bind().Body(&req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).
			JSON(dto.CreateResponseError(err.Error()))
	}
	fails := util.Validate(req)
	if len(fails) > 0 {
		return ctx.Status(http.StatusBadRequest).JSON(dto.CreateResponseErrorData("Validation failed", fails))
	}

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(errorStatus(err)).JSON(dto.CreateResponseError(err.Error()))
	}
	req.Id = id.String()
	userId, err := getUserId(ctx)
	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}
	req.UserId = userId
	err = ca.cartService.UpdateItemQuantity(c, req)
	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}

	return ctx.SendStatus(http.StatusOK)
}
func (ca cartApi) DeleteItem(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(errorStatus(err)).JSON(dto.CreateResponseError(err.Error()))
	}
	userId, err := getUserId(ctx)
	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}
	err = ca.cartService.DeleteItem(c, id.String(), userId)

	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}

	return ctx.SendStatus(http.StatusNoContent)

}
