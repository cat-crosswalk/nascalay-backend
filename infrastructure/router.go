package infrastructure

import (
	"log"

	"github.com/21hack02win/nascalay-backend/interfaces/handler/oapi"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(e *echo.Echo) {
	mw, err := oapi.SetupOapiMiddleware()
	if err != nil {
		log.Fatal(err)
	}

	e.Use(
		middleware.Logger(),
		middleware.Recover(),
		mw,
	)

	s := injectServer()
	oapi.RegisterHandlers(e, s)
}
