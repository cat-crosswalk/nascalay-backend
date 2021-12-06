package handler

import (
	"net/http"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/labstack/echo/v4"
)

func (h *handler) Ws(c echo.Context, params WsParams) error {
	err := h.stream.ServeWS(c.Response().Writer, c.Request(), model.UserId(params.NascalayUser))
	if err != nil {
		c.Logger().Error(err)
		return newEchoHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
