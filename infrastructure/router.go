package infrastructure

import (
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(e *echo.Echo, endpoint string) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	s := injectServer()
	oapi.RegisterHandlersWithBaseURL(e, s, endpoint)
}
