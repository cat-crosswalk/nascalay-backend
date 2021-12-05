package model

import "github.com/gofrs/uuid"

type Room struct {
	Id       RoomId
	Capacity int
	HostId   uuid.UUID
	Members  []User
}

type RoomId string

func (rid RoomId) String() string {
	return string(rid)
}
