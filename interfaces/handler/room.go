package handler

import (
	"net/http"

	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/labstack/echo/v4"
)

func (h *handler) JoinRoom(c echo.Context) error {
	return nil
}

func (h *handler) CreateRoom(c echo.Context) error {
	req := new(CreateRoomJSONRequestBody)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err) // TODO: エラーそのまま返すのどうにかする
	}

	room, err := h.r.CreateRoom(&repository.CreateRoomArgs{
		Avatar:   req.Avatar,
		Capacity: req.Capacity,
		Username: req.Username,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return echo.NewHTTPError(http.StatusCreated, refillRoom(room, room.HostId))
}

func (h *handler) GetRoom(c echo.Context, roomId RoomId) error {
	return nil
}
