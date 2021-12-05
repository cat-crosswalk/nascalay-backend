package handler

import (
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/gofrs/uuid"
)

func refillRoom(mr *model.Room, userId uuid.UUID) Room {
	var r Room
	r.Capacity = mr.Capacity
	r.HostId = mr.HostId
	r.Members = make([]User, len(mr.Members))
	r.RoomId = mr.Id
	r.UserId = userId

	for i, v := range mr.Members {
		r.Members[i] = refillUser(&v)
	}

	return r
}

func refillUser(mu *model.User) User {
	var u User
	u.Avatar = mu.Avatar
	u.UserId = mu.Id
	u.Username = mu.Name

	return u
}
