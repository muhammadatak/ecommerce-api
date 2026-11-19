package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/muhammadatak/go-ecommerce-api/domain"
	"github.com/muhammadatak/go-ecommerce-api/dto"
)

type productApi struct {
	productService domain.ProductService
}

func NewProduct(app *fiber.App, productService domain.ProductService, authMid fiber.Handler) {
	pa := productApi{
		productService: productService,
	}

	products := app.Group("/products", authMid)

	products.Get("/", pa.Index)
	products.Get("/:id", pa.Show)
	products.Post("/", pa.Create)
	products.Patch("/:id", pa.Update)
	products.Delete("/:id", pa.Delete)

}

func (pa productApi) Index(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	res, err := pa.productService.Index(c)

	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}
	return ctx.JSON(dto.CreateResponseSuccess(res))
}

func (pa productApi) Create(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	var req dto.CreateProductReq

	err := ctx.Bind().Body(&req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).
			JSON(dto.CreateResponseError(err.Error()))
	}

	res, err := pa.productService.Create(c, req)

	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}
	return ctx.JSON(dto.CreateResponseSuccess(res))
}

func (pa productApi) Show(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).
			JSON(dto.CreateResponseError(err.Error()))
	}
	res, err := pa.productService.Show(c, id.String())

	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}
	return ctx.JSON(dto.CreateResponseSuccess(res))
}

func (pa productApi) Update(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	var req dto.UpdateProductReq

	err := ctx.Bind().Body(&req)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).
			JSON(dto.CreateResponseError(err.Error()))
	}

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).
			JSON(dto.CreateResponseError(err.Error()))
	}
	req.Id = id.String()
	res, err := pa.productService.Update(c, req)

	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}
	return ctx.JSON(dto.CreateResponseSuccess(res))
}

func (pa productApi) Delete(ctx fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).
			JSON(dto.CreateResponseError(err.Error()))
	}
	err = pa.productService.Delete(c, id.String())

	if err != nil {
		return ctx.Status(errorStatus(err)).
			JSON(dto.CreateResponseError(err.Error()))
	}

	return ctx.SendStatus(http.StatusNoContent)
}
