package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type orderApi struct {
	orderService domain.OrderService
}

func NewOrder(app *fiber.App, orderService domain.OrderService, authMid fiber.Handler) {
	oa := orderApi{
		orderService: orderService,
	}

	orders := app.Group("/orders", authMid)

	orders.Post("/", oa.Checkout)
}

func (oa orderApi) Checkout(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	var req dto.CheckoutRequest

	userId, err := getUserId(ctx)
	if err != nil {
		return ctx.Status(errorStatus(err)).JSON(dto.CreateResponseError(err.Error()))
	}

	req.UserId = userId

	res, err := oa.orderService.Checkout(c, req)

	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}

	return ctx.Status(http.StatusCreated).JSON(dto.CreateResponseSuccess(res))
}
