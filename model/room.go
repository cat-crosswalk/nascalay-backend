package model

import "github.com/gofrs/uuid"

type Room struct {
	RoomId   string
	UserId   uuid.UUID
	Name     string
	Capacity int
}
