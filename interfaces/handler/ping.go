package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *handler) Ping(c echo.Context) error {
	return echo.NewHTTPError(http.StatusOK, "pong")
}
