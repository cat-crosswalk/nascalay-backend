package infrastructure

import (
	"strings"

	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(e *echo.Echo, baseEndpoint string) {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.String(), "/assets")
		},
	}))
	e.Use(middleware.Recover())

	s := injectServer()
	oapi.RegisterHandlersWithBaseURL(e, s, baseEndpoint)
}
