package model

import "github.com/gofrs/uuid"

type User struct {
	Id     uuid.UUID
	Name   string
	Avatar int
}
