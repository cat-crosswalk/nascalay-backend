package handler

import (
	"net/http"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

func (h *handler) Ws(c echo.Context) error {
	userId := c.Request().Header.Get("Nascalay-User")

	err := h.stream.ServeWS(c.Response().Writer, c.Request(), model.UserId(uuid.FromStringOrNil(userId)))
	if err != nil {
		c.Logger().Error(err)
		return err // TODO: あとで
	}

	return c.NoContent(http.StatusNoContent)
}
