package model

import "github.com/gofrs/uuid"

type User struct {
	Id     UserId
	Name   string
	Avatar int
}

type UserId uuid.UUID

func (uid UserId) UUID() uuid.UUID {
	return uuid.UUID(uid)
}
