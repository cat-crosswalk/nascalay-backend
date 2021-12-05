package model

import "github.com/gofrs/uuid"

type Room struct {
	Id       string
	Capacity int
	HostId   uuid.UUID
	Members  []User
}
