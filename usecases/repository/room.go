package repository

import (
	"github.com/21hack02win/nascalay-backend/model"
)

type RoomRepository interface {
	JoinRoom(jr *JoinRoomArgs) (*model.Room, model.UserId, error)
	CreateRoom(cr *CreateRoomArgs) (*model.Room, error)
}

type CreateRoomArgs struct {
	Avatar   model.Avatar
	Capacity int
	Username model.Username
}

type JoinRoomArgs struct {
	Avatar   model.Avatar
	RoomId   model.RoomId
	Username model.Username
}
