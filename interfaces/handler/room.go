package handler

import (
	"net/http"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

func (h *handler) JoinRoom(c echo.Context) error {
	req := new(JoinRoomJSONRequestBody)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	room, uid, err := h.r.JoinRoom(&repository.JoinRoomArgs{
		Avatar:   model.Avatar(req.Avatar),
		RoomId:   model.RoomId(req.RoomId),
		Username: model.Username(req.Username),
	})
	if err != nil {
		return newEchoHTTPError(err)
	}

	return echo.NewHTTPError(http.StatusOK, refillRoom(room, uid))
}

func (h *handler) CreateRoom(c echo.Context) error {
	req := new(CreateRoomJSONRequestBody)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	room, err := h.r.CreateRoom(&repository.CreateRoomArgs{
		Avatar:   model.Avatar(req.Avatar),
		Capacity: model.Capacity(req.Capacity),
		Username: model.Username(req.Username),
	})
	if err != nil {
		return newEchoHTTPError(err)
	}

	return echo.NewHTTPError(http.StatusCreated, refillRoom(room, room.HostId))
}

func (h *handler) GetRoom(c echo.Context, roomId RoomIdInPath) error {
	room, err := h.r.GetRoom(string(roomId))
	if err != nil {
		return newEchoHTTPError(err)
	}

	return c.JSON(http.StatusOK, refillRoom(room, model.UserId(uuid.Nil))) // ユーザーIDが必要ないのでとりあえずuuid.Nilにしておく
}
