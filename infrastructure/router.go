package infrastructure

import (
	"github.com/21hack02win/nascalay-backend/interfaces/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(e *echo.Echo) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	s := injectServer()
	handler.RegisterHandlers(e, s)
}
