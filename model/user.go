package model

import "github.com/gofrs/uuid"

type User struct {
	Id     UserId
	Name   Username
	Avatar int
}

type UserId uuid.UUID

func (uid UserId) UUID() uuid.UUID {
	return uuid.UUID(uid)
}

type Username string

func (un Username) String() string {
	return string(un)
}
