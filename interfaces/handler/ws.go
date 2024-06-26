package handler

import (
	"errors"
	"net/http"

	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/labstack/echo/v4"
)

func (h *handler) Ws(c echo.Context, params oapi.WsParams) error {
	uid, err := params.User.Refill()
	if err != nil {
		// Invalid uuid
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = h.ws.ServeWS(c.Response().Writer, c.Request(), uid)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return newEchoHTTPError(err, c)
	}

	return c.NoContent(http.StatusNoContent)
}
