package handler

import (
	"net/http"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/21hack02win/nascalay-backend/util/session"
	"github.com/labstack/echo/v4"
)

func (h *handler) JoinRoom(c echo.Context) error {
	return nil
}

func (h *handler) CreateRoom(c echo.Context) error {
	req := new(CreateRoomJSONRequestBody)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	room, err := h.r.CreateRoom(&repository.CreateRoomArgs{
		Avatar:   req.Avatar,
		Capacity: req.Capacity,
		Username: req.Username,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}

	if err := session.Set(session.UserIdKey, model.User{
		Id:     room.HostId,
		Name:   req.Username,
		Avatar: req.Avatar,
	}, c); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return echo.NewHTTPError(http.StatusCreated, refillRoom(room, room.HostId))
}

func (h *handler) GetRoom(c echo.Context, roomId RoomId) error {
	return nil
}
