package handler

import (
	"net/http"

	"github.com/21hack02win/nascalay-backend/interfaces/handler/oapi"
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

func (h *handler) Ws(c echo.Context, params oapi.WsParams) error {
	if err := h.stream.ServeWS(c.Response().Writer, c.Request(), model.UserId(uuid.FromStringOrNil(string(params.User)))); err != nil {
		c.Logger().Error(err)
		return newEchoHTTPError(err)
	}

	return echo.NewHTTPError(http.StatusNoContent)
}
