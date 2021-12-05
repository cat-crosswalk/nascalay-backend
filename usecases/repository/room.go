package repository

import (
	"github.com/21hack02win/nascalay-backend/model"
)

type RoomRepository interface {
	CreateRoom(cr *CreateRoomArgs) (*model.Room, error)
}

type CreateRoomArgs struct {
	Name     string
	Capacity int
}
