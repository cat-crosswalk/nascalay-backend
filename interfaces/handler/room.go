package handler

import (
	"fmt"
	"net/http"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/labstack/echo/v4"
)

func (h *handler) JoinRoom(c echo.Context) error {
	req := new(oapi.JoinRoomJSONRequestBody)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	room, uid, err := h.r.JoinRoom(&repository.JoinRoomArgs{
		Avatar: model.Avatar{
			Type:  model.AvatarType(req.Avatar.Type),
			Color: model.AvatarColor(req.Avatar.Color),
		},
		RoomId:   model.RoomId(req.RoomId),
		Username: model.Username(req.Username),
	})
	if err != nil {
		return newEchoHTTPError(err, c)
	}

	// Notify Other Clients of the new user with WebSocket
	if err := h.stream.NotifyOfNewRoomMember(room); err != nil {
		c.Logger().Error(fmt.Errorf("failed to notify of new member: %w", err))
	}

	return c.JSON(http.StatusOK, oapi.RefillRoom(room, uid))
}

func (h *handler) CreateRoom(c echo.Context) error {
	req := new(oapi.CreateRoomJSONRequestBody)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	room, err := h.r.CreateRoom(&repository.CreateRoomArgs{
		Avatar: model.Avatar{
			Type:  model.AvatarType(req.Avatar.Type),
			Color: model.AvatarColor(req.Avatar.Color),
		},
		Capacity: model.Capacity(req.Capacity),
		Username: model.Username(req.Username),
	})
	if err != nil {
		return newEchoHTTPError(err, c)
	}

	return c.JSON(http.StatusCreated, oapi.RefillRoom(room, room.HostId))
}

func (h *handler) GetRoom(c echo.Context, roomId oapi.RoomIdInPath) error {
	room, err := h.r.GetRoom(model.RoomId(roomId))
	if err != nil {
		return newEchoHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, oapi.RefillRoom(room, model.UserId{})) // ユーザーIDが必要ないのでとりあえずuuid.Nilにしておく
}
