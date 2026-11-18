package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
	"github.com/muhammadatak/go-ecommerce-api/internal/util"
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

	err = oa.orderService.Checkout(c, req)

	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}

	return ctx.SendStatus(http.StatusCreated)
}
