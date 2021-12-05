package repository

import (
	"github.com/21hack02win/nascalay-backend/model"
)

type RoomRepository interface {
	JoinRoom(jr *JoinRoomArgs) (*model.Room, model.UserId, error)
	CreateRoom(cr *CreateRoomArgs) (*model.Room, error)
}

type CreateRoomArgs struct {
	Avatar   int
	Capacity int
	Username string
}

type JoinRoomArgs struct {
	Avatar   int
	RoomId   model.RoomId
	Username string
}
