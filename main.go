package main

import (
	"net/http"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/muhammadatak/go-ecommerce-api/dto"
	"github.com/muhammadatak/go-ecommerce-api/internal/api"
	"github.com/muhammadatak/go-ecommerce-api/internal/config"
	"github.com/muhammadatak/go-ecommerce-api/internal/connection"
	"github.com/muhammadatak/go-ecommerce-api/internal/repository"
	"github.com/muhammadatak/go-ecommerce-api/internal/service"
)

func main() {
	cnf := config.Get()
	dbConn := connection.GetDatabase(cnf.Database)
	jwtMidd := jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(cnf.Jwt.Key)},
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			return ctx.Status(http.StatusUnauthorized).
				JSON(dto.CreateResponseError("endpoint perlu token, silakan login dulu"))
		},
	})

	app := fiber.New()
	//repository
	userRepos := repository.NewUser(dbConn)
	productRepository := repository.NewProduct(dbConn)

	//service
	authRepos := service.NewAuth(userRepos, cnf)
	productService := service.NewProduct(productRepository)

	//api
	api.NewProduct(app, productService, jwtMidd)
	api.NewAuth(app, authRepos)

	_ = app.Listen(cnf.Server.Host + ":" + cnf.Server.Port)

}
