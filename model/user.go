package model

import (
	"fmt"

	"github.com/gofrs/uuid"
)

type User struct {
	Id     UserId
	Name   Username
	Avatar Avatar
}

type UserId uuid.UUID

func (uid UserId) UUID() uuid.UUID {
	return uuid.UUID(uid)
}

func UserIdFromString(str string) (UserId, error) {
	uid, err := uuid.FromString(str)
	if err != nil {
		return UserId{}, fmt.Errorf("failed to convert UserId: %w", err)
	}

	return UserId(uid), nil
}

type Username string

func (un Username) String() string {
	return string(un)
}

type Avatar int

func (a Avatar) Int() int {
	return int(a)
}

const (
	Avatar0 Avatar = iota // Note: ここにAvatarの種類を追記する
	AvatarLimit
)
