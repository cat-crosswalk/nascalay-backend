package ws

import (
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/labstack/echo/v4"
)

type RoomServer struct {
	hub    *Hub
	room   *model.Room
	logger echo.Logger
}
