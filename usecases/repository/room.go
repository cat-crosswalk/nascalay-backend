package repository

import (
	"github.com/21hack02win/nascalay-backend/model"
)

type RoomRepository interface {
	JoinRoom(jr *JoinRoomArgs) (*model.Room, model.UserId, error)
	CreateRoom(cr *CreateRoomArgs) (*model.Room, error)
	GetRoom(rid model.RoomId) (*model.Room, error)
	GetRoomFromUserId(uid model.UserId) (*model.Room, error)
}

type CreateRoomArgs struct {
	Avatar   model.Avatar
	Capacity model.Capacity
	Username model.Username
}

type JoinRoomArgs struct {
	Avatar   model.Avatar
	RoomId   model.RoomId
	Username model.Username
}
