package handler

import (
	"net/http"

	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/labstack/echo/v4"
)

func (h *handler) Ws(c echo.Context, params oapi.WsParams) error {
	uid, err := params.User.Refill()
	if err != nil {
		// Invalid uuid
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.stream.ServeWS(c.Response().Writer, c.Request(), uid); err != nil {
		return newEchoHTTPError(err, c)
	}

	return c.NoContent(http.StatusNoContent)
}
