package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *handler) Ping(c echo.Context) error {
	return c.JSON(http.StatusOK, "pong")
}
